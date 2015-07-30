package websocket

import (
	"encoding/json"
	"sync"
	"time"

	log "github.com/cihub/seelog"
	"github.com/go-martini/martini"
	"github.com/gorilla/websocket"
)

type Message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
	To   string          `json:"to"`
}
type Clients struct {
	sync.Mutex
	clients []*Client
	Router  *Router `inject:""`
}
type Client struct {
	Name       string
	in         <-chan *Message
	out        chan<- *Message
	done       <-chan bool
	err        <-chan error
	disconnect chan<- int
}

// Add a client to a room
func (r *Clients) appendClient(client *Client) {
	r.Lock()
	r.clients = append(r.clients, client)
	r.Unlock()

	msgs := r.Router.RunOnClientConnectHandlers()
	for _, msg := range msgs {
		client.out <- msg
	}
}

// Message all the other clients
func (r *Clients) SendToAll(t string, data interface{}) {

	out, err := json.Marshal(data)
	if err != nil {
		log.Error(err)
		return
	}
	msg := &Message{Type: t, Data: out}

	r.Lock()
	defer r.Unlock()
	clientsToRemove := make([]*Client, 0)

	for _, c := range r.clients {
		select {
		case c.out <- msg:
			// Everything went well :)
		case <-time.After(time.Second):
			log.Warn("Failed writing to websocket: timeout (", c.Name, ")")
			clientsToRemove = append(clientsToRemove, c)
		}
	}

	go func() {
		for _, c := range clientsToRemove {
			r.removeClient(c)
		}
	}()
}

// Remove a client
func (r *Clients) removeClient(client *Client) {
	r.Lock()
	defer r.Unlock()

	for index, c := range r.clients {
		if c == client {
			c.disconnect <- websocket.CloseInternalServerErr
			r.clients = append(r.clients[:index], r.clients[(index+1):]...)
		}
	}
}

// Disconnect all clients
func (r *Clients) disconnectAll() {
	r.Lock()
	defer r.Unlock()

	for _, c := range r.clients {
		c.disconnect <- websocket.CloseGoingAway
	}
}

func newClients() *Clients {
	return &Clients{sync.Mutex{}, make([]*Client, 0), nil}
}
func (clients *Clients) WebsocketRoute(params martini.Params, receiver <-chan *Message, sender chan<- *Message, done <-chan bool, disconnect chan<- int, err <-chan error) (int, string) {
	client := &Client{params["clientname"], receiver, sender, done, err, disconnect}
	clients.appendClient(client)

	// A single select can be used to do all the messaging
	for {
		select {
		case <-client.err:
			// Don't try to do this:
			// client.out <- &Message{"system", "system", "There has been an error with your connection"}
			// The socket connection is already long gone.
			// Use the error for statistics etc
		case msg := <-client.in:
			//TODO implement command from websocket here. using same process as WebHandlerCommandToNode

			log.Info("incoming message from webui on websocket", string(msg.Data))
			clients.Router.Run(msg)
		case <-client.done:
			clients.removeClient(client)
			return 200, "OK"
		}
	}
}
