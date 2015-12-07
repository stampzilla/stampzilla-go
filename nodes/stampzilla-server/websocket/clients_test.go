package websocket

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAppendClient(t *testing.T) {

	clients := newClients()
	clients.Router = NewRouter()

	clients.Router.AddClientConnectHandler(func() *Message {
		return &Message{
			Type: "type",
		}
	})

	client := &Client{}
	client.out = make(chan *Message)

	go func() {
		select {
		case msg := <-client.out:
			assert.Equal(t, msg.Type, "type")
			close(client.out)
		case <-time.After(time.Second):

		}
	}()

	clients.appendClient(client)

}
func TestSendToAll(t *testing.T) {

	clients := newClients()
	clients.Router = NewRouter()

	clients.Router.AddClientConnectHandler(func() *Message {
		return &Message{
			Type: "type",
		}
	})

	client := &Client{}
	client.out = make(chan *Message)
	go clients.appendClient(client)
	<-client.out

	go func() {
		for {
			select {
			case msg := <-client.out:
				assert.Equal(t, "test", msg.Type)
				assert.Equal(t, `"data"`, string(msg.Data))
			case <-time.After(time.Second):

			}
		}
	}()

	clients.SendToAll("test", "data")

}
