package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/mmfKupl/gosse"
)

var notifier gosse.INotifier

func init() {
	log.SetFlags(log.Llongfile + log.Ldate + log.Ltime)
}

func main() {
	notifier = initNotifier()

	http.Handle("/connect", notifier)
	http.HandleFunc("/notify", notifyHandler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		pwd, _ := os.Getwd()
		file, err := ioutil.ReadFile(path.Join(pwd, "examples/notifier-page.html"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, _ = w.Write(file)
	})

	go func() {
		for i := 0; i < 100; i++ {
			time.Sleep(1 * time.Second)
			notifier.NotifyAll(gosse.Message(fmt.Sprintf("%v", i)))
		}
	}()

	log.Printf("Starting server at port 8080\n")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func notifyHandler(w http.ResponseWriter, r *http.Request) {
	keys := r.URL.Query()
	clientId := keys.Get("clientId")
	connectionId := keys.Get("connectionId")
	message := keys.Get("message")
	if message == "" {
		message = "default message"
	}

	if clientId == "" {
		notifier.NotifyAll(gosse.Message(message))
		_, _ = fmt.Fprintf(w, "All clients successfuly notified, message - `%s`.", message)
		return
	}

	client := notifier.GetClient(clientId)
	if client == nil {
		http.Error(w, fmt.Sprintf("Client with id = `%s` dosn't fond.", clientId), http.StatusBadRequest)
		return
	}

	if connectionId == "" {
		client.Notify(gosse.Message(message))
		_, _ = fmt.Fprintf(w, "Successfully notified all connection on client with id = `%s, message - `%s`.", clientId, message)
		return
	}

	err := client.NotifyConnection(connectionId, gosse.Message(message))
	if err != nil {
		http.Error(w, fmt.Sprintf("Cannot notify client with id = `%s`, on connection with id = `%s`, error - %s.", clientId, connectionId, err), http.StatusInternalServerError)
		return
	}
	_, _ = fmt.Fprintf(w, "Successfully notified connection with id = `%s` on client with id = `%s, message - `%s`.", connectionId, clientId, message)
}

func initNotifier() gosse.INotifier {
	notifier := &gosse.Notifier{}
	notifier.Init()
	notifier.RegisterClientIdentifier(clientIdentifier)
	return notifier
}

func clientIdentifier(r *http.Request) (string, error) {
	clientId := r.URL.Query().Get("clientId")
	if clientId == "" {
		return "", fmt.Errorf("Can not get valid id. ")
	}
	return clientId, nil
}
