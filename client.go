package gosse

import "fmt"

// IClient - интерфейс, описывающий методы,
// необходимые для работы с клиентом и его подключениями
type IClient interface {
	RegisterConnection(id string, connection IConnection) error
	RemoveConnection(id string)
	CloseAndRemoveConnection(id string)
	NotifyConnection(id string, message Message) error
	Notify(message Message)
	GetId() string
	GetConnection(id string) IConnection
}

// Client - структура клиента, хранит све подключения,
// имплементирует работу с подключениями
type Client struct {
	id          string
	connections map[string]IConnection
}

func (c *Client) GetId() string {
	return c.id
}

func (c *Client) RegisterConnection(id string, connection IConnection) error {
	if c.connections[id] != nil {
		return fmt.Errorf("connection with id = `%s` already registred", id)
	}
	c.connections[id] = connection
	return nil
}

func (c *Client) RemoveConnection(id string) {
	delete(c.connections, id)
}

func (c *Client) CloseAndRemoveConnection(id string) {
	if c.connections[id] != nil {
		c.connections[id].Close()
	}
	delete(c.connections, id)
}

func (c *Client) NotifyConnection(id string, message Message) error {
	if connection, ok := c.connections[id]; ok {
		connection.Notify(message)
		return nil
	}
	return fmt.Errorf("connection with id = `%s` dosn't registred", id)
}

func (c *Client) Notify(message Message) {
	for _, connection := range c.connections {
		connection.Notify(message)
	}
}

func (c *Client) GetConnection(id string) IConnection {
	return c.connections[id]
}
