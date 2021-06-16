package gosse

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

// INotifier - интерфейс, описывающий методы, необходимые для работы с клиентами
type INotifier interface {
	RegisterClient(id string, client IClient) error
	RemoveClient(id string)
	GetClient(id string) IClient
	NotifyAll(message Message)
	RegisterClientIdentifier(fn func(r *http.Request) (string, error))
	IdentifyClient(r *http.Request) (string, error)
	RegisterConnectionIdentifier(fn func(r *http.Request) (string, error))
	IdentifyConnection(r *http.Request) (string, error)
	RegisterOnClientRegistered(fn func(c IClient))
	OnClientRegistered(c IClient)
	RegisterOnConnectionRegistered(fn func(c IClient, conn IConnection))
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

func defaultClientIdentifier() string {
	return uuid.New().String()
}

func defaultConnectionIdentifier() string {
	return uuid.New().String()
}

// Notifier - структура, хранящая клиентов и имплементирующая работу с ними
type Notifier struct {
	clients                map[string]IClient
	clientIdentifier       func(r *http.Request) (string, error)
	connectionIdentifier   func(r *http.Request) (string, error)
	onClientRegistered     func(client IClient)
	onConnectionRegistered func(c IClient, conn IConnection)
}

func (n *Notifier) Init() {
	n.clients = make(map[string]IClient)
}

func (n *Notifier) RegisterClient(id string, client IClient) error {
	if n.clients[id] != nil {
		return fmt.Errorf("client with id = `%s` already registred", id)
	}
	n.clients[id] = client
	go n.OnClientRegistered(client)
	return nil
}

func (n *Notifier) RemoveClient(id string) {
	delete(n.clients, id)
}

func (n *Notifier) GetClient(id string) IClient {
	return n.clients[id]
}

func (n *Notifier) NotifyAll(message Message) {
	for _, client := range n.clients {
		client.Notify(message)
	}
}

func (n *Notifier) RegisterClientIdentifier(fn func(r *http.Request) (string, error)) {
	n.clientIdentifier = fn
}

func (n *Notifier) IdentifyClient(r *http.Request) (string, error) {
	if n.clientIdentifier == nil {
		return defaultClientIdentifier(), nil
	}
	return n.clientIdentifier(r)
}

func (n *Notifier) RegisterConnectionIdentifier(fn func(r *http.Request) (string, error)) {
	n.connectionIdentifier = fn
}

func (n *Notifier) IdentifyConnection(r *http.Request) (string, error) {
	if n.connectionIdentifier == nil {
		return defaultConnectionIdentifier(), nil
	}
	return n.connectionIdentifier(r)
}

func (n *Notifier) getIdentifiedClient(r *http.Request) (IClient, error) {
	clientId, err := n.IdentifyClient(r)
	if err != nil {
		return nil, fmt.Errorf("Identifying client failed: %s. ", err)
	}

	client := n.GetClient(clientId)

	if client == nil {
		client = &Client{
			connections: make(map[string]IConnection),
			id: clientId,
			onConnectionRegistered: n.onConnectionRegistered,
		}
		if err = n.RegisterClient(clientId, client); err != nil {
			return nil, fmt.Errorf("Register client with id = `%s` failed: %s. ", clientId, err)
		}
	}

	return client, nil
}

func (n *Notifier) getIdentifiedConnection(client IClient, r *http.Request) (IConnection, error) {
	connectionId, err := n.IdentifyConnection(r)
	if err != nil {
		return nil, fmt.Errorf("Identifying connection failed: %s. ", err)
	}

	connection := client.GetConnection(connectionId)
	if connection == nil {
		connection = &Connection{messageChan: make(chan Message), id: connectionId}
		if err = client.RegisterConnection(connectionId, connection); err != nil {
			return nil, fmt.Errorf("Register connection with id = `%s` failed: %s. ", connectionId, err)
		}
	}

	return connection, nil
}

func (n *Notifier) RegisterOnClientRegistered(fn func(c IClient)) {
	n.onClientRegistered = fn
}

func (n *Notifier) OnClientRegistered(client IClient) {
	if n.onClientRegistered != nil {
		n.onClientRegistered(client)
	}
}

func (n *Notifier) RegisterOnConnectionRegistered(fn func(c IClient, conn IConnection)) {
	n.onConnectionRegistered = fn

	for _, client := range n.clients {
		client.RegisterOnConnectionRegistered(n.onConnectionRegistered)
	}
}

func setStreamingHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
}

func (n *Notifier) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)

	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	setStreamingHeaders(w)

	client, err := n.getIdentifiedClient(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("Getting client failed: %s", err), http.StatusInternalServerError)
		return
	}

	connection, err := n.getIdentifiedConnection(client, r)
	if err != nil {
		http.Error(w, fmt.Sprintf("Getting connection failed: %s", err), http.StatusInternalServerError)
		return
	}

	defer func() {
		client.CloseAndRemoveConnection(connection.GetId())
	}()

	doneChan := r.Context().Done()

	go func() {
		<-doneChan
		client.CloseAndRemoveConnection(connection.GetId())
	}()

	for {
		message, ok := connection.ReadMessage()
		if !ok {
			break
		}
		_, err = fmt.Fprintf(w, "data:%s\n\n", message)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error acured: %s", err), http.StatusGone)
			break
		}
		flusher.Flush()
	}

}
