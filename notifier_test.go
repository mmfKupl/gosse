package gosse

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func TestINotifierCreatingAndNotifying(t *testing.T) {

	var notifier INotifier = &Notifier{clients: make(map[string]IClient)}

	var client IClient = &Client{connections: make(map[string]IConnection)}
	clientId := "firstClient"

	if err := notifier.RegisterClient(clientId, client); err != nil {
		t.Errorf("Register client with id = `%s` failed, error = `%s`", clientId, err)
	} else {
		t.Logf("Register client with id = `%s` successfully", clientId)
	}

	expectedFirstResult := "0123456789first"
	firstConnectionId := "fistConnectionId"
	var firstConnection IConnection = &Connection{messageChan: make(chan Message)}
	if err := client.RegisterConnection(firstConnectionId, firstConnection); err != nil {
		t.Errorf("Register connection error - %s", err)
		t.FailNow()
	} else {
		t.Logf("Connection with id = `%s` registered successfully", firstConnectionId)
	}
	if err := client.RegisterConnection(firstConnectionId, firstConnection); err == nil {
		t.Errorf("Register second connection with id = `%s` - error dosn't recived", firstConnectionId)
	} else {
		t.Logf("Register second connection with id = `%s` - error recived successfully - `%s`", firstConnectionId, err)
	}

	go func() {
		for i := 0; i < 10; i++ {
			client.Notify(Message(strconv.Itoa(i)))
		}
		err := client.NotifyConnection(firstConnectionId, Message("first"))
		if err != nil {
			t.Errorf("Error with notifying client: %s", err)
		}
		client.CloseAndRemoveConnection(firstConnectionId)
	}()

	firstResult := listenConnection(firstConnection)

	if firstResult != expectedFirstResult {
		t.Errorf(
			"Received message of connection with id = `%s` - `%s` dosn't equal sendet message `%s`\n",
			firstConnectionId,
			firstResult,
			expectedFirstResult,
		)
	} else {
		t.Logf("Recived message of connection with id = `%s` successfully", firstConnectionId)
	}
}

func TestNotifier_ServeHTTP(t *testing.T) {

	expectedResult := "0123456789"

	notifier := &Notifier{
		clients: make(map[string]IClient),
	}

	clientIdentifier := func(r *http.Request) (string, error) {
		return "clientId-1", nil
	}
	connectionIdentifier := func(r *http.Request) (string, error) {
		return "connectionId-1", nil
	}

	notifier.RegisterClientIdentifier(clientIdentifier)
	notifier.RegisterConnectionIdentifier(connectionIdentifier)

	testServer := httptest.NewServer(http.HandlerFunc(notifier.ServeHTTP))
	defer testServer.Close()

	closeChan := make(chan struct{})

	var resp *http.Response
	var err error

	notifyAll := make(chan struct{})

	go func() {
		time.Sleep(1 * time.Second)
		for i := 0; i < 10; i++ {
			notifier.NotifyAll(Message(strconv.Itoa(i)))
		}

		client := notifier.GetClient("clientId-1")
		client.CloseAndRemoveConnection("connectionId-1")
		closeChan <- struct{}{}
		notifyAll <- struct{}{}
		close(notifyAll)
	}()

	go func() {
		resp, err = http.Get(testServer.URL)
		if err != nil {
			t.Errorf("Cannot get to testServer to url `%s` with notifier - `%s`", testServer.URL, err)
			t.FailNow()
		}

		<-notifyAll
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("Reading response body failed: %e", err)
			t.FailNow()
		}
		bodyString := string(bodyBytes)
		if bodyString != expectedResult {
			t.Errorf("Notifying data (`%s`) dosn't equal sendet (`%s`)", bodyString, expectedResult)
		}

		closeChan <- struct{}{}
	}()

	i := 0
	for {
		<-closeChan
		if i++; i == 2 {
			close(closeChan)
			break
		}
	}

	testServer.Close()
}
