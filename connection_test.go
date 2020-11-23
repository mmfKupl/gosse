package gosse

import (
	"reflect"
	"strconv"
	"testing"
)

func listenConnection(connection IConnection) string {
	result := ""
	for {
		message, ok := connection.ReadMessage()
		if ok {
			result += string(message)
		} else {
			break
		}
	}
	return result
}

func TestIConnectionMessagesReading(t *testing.T) {
	var connection IConnection = &Connection{messageChan: make(chan Message)}
	expectedResult := "0123456789"

	go func() {
		for i := 0; i < 10; i++ {
			connection.Notify(Message(strconv.Itoa(i)))
		}
		connection.Close()
	}()

	var chanEntity interface{} = connection.GetChan()
	if _, ok := chanEntity.(chan Message); !ok {
		t.Errorf("Connection.GetChan() returns not `chan Message` type - %s\n", reflect.TypeOf(chanEntity))
		t.FailNow()
	} else {
		t.Logf("Creating chan type valid")
	}

	receivedMessage := listenConnection(connection)

	if receivedMessage != expectedResult {
		t.Errorf("Received message `%s` dosn't equal sendet message - `%s`\n", receivedMessage, expectedResult)
	} else {
		t.Logf("Recived message valid")
	}

}
