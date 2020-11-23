package gosse

// Message - это алиас []byte. Используется для конкретизации в пакете
type Message []byte

// IConnection - интерфейс, описывающий методы,
// необходимые для работы с подключением
type IConnection interface {
	Notify(message Message)
	ReadMessage() (Message, bool)
	Close()
	GetChan() chan Message
	GetId() string
}

// Connection - стркутура соединения, содержит в себе экземпляр канала Message
type Connection struct {
	id          string
	messageChan chan Message
}

func (c *Connection) GetId() string {
	return c.id
}

func (c *Connection) Notify(message Message) {
	c.messageChan <- message
}

func (c *Connection) ReadMessage() (Message, bool) {
	message, ok := <-c.messageChan
	return message, ok
}

func (c *Connection) Close() {
	close(c.messageChan)
}

func (c *Connection) GetChan() chan Message {
	return c.messageChan
}
