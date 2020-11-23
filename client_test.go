package gosse

import (
	"strconv"
	"testing"
)

func TestIClientCreatingAndNotifyingClient(t *testing.T) {
	var client IClient = &Client{connections: make(map[string]IConnection)}

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

	expectedSecondResult := "0123456789second"
	secondConnectionId := "secondConnectionId"
	var secondConnection IConnection = &Connection{messageChan: make(chan Message)}
	if err := client.RegisterConnection(secondConnectionId, secondConnection); err != nil {
		t.Errorf("Register connection error - %s", err)
		t.FailNow()
	} else {
		t.Logf("Connection with id = `%s` registered successfully", secondConnectionId)
	}

	go func() {
		for i := 0; i < 10; i++ {
			client.Notify(Message(strconv.Itoa(i)))
		}
		err := client.NotifyConnection(firstConnectionId, Message("first"))
		if err != nil {
			t.Errorf("Error with notifying client: %s", err)
		}
		err = nil
		err = client.NotifyConnection(secondConnectionId, Message("second"))
		if err != nil {
			t.Errorf("Error with notifying client: %s", err)
		}

		client.CloseAndRemoveConnection(firstConnectionId)
		client.CloseAndRemoveConnection(secondConnectionId)
	}()

	endingChan := make(chan struct{})

	firstResult := ""
	go func() {
		firstResult = listenConnection(firstConnection)
		endingChan <- struct{}{}
	}()

	secondResult := ""
	go func() {
		secondResult = listenConnection(secondConnection)
		endingChan <- struct{}{}
	}()

	i := 0
	for range endingChan {
		if i++; i == 2 {
			break
		}
	}

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

	if secondResult != expectedSecondResult {
		t.Errorf(
			"Received message of connection with id = %s - `%s` dosn't equal sendet message `%s`\n",
			secondConnectionId,
			secondResult,
			expectedSecondResult,
		)
	} else {
		t.Logf("Recived message of connection with id = `%s` successfully", secondConnectionId)
	}

}
