package websocket

import (
	"encoding/json"
	"fmt"
	"sync"

	log "github.com/cihub/seelog"
	"github.com/go-martini/martini"
	"github.com/gorilla/websocket"
)

type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
	To   string      `json:"to"`
}
type Clients struct {
	sync.Mutex
	clients []*Client
	Router  *Router `inject:""`
}
type Client struct {
	Name       string
	in         <-chan string
	out        chan<- string
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
		str, err := json.Marshal(msg)
		if err != nil {
			log.Error(err)
		}
		client.out <- string(str)
	}
}

// Message all the other clients
func (r *Clients) SendToAll(msg interface{}) {

	var str string
	switch t := msg.(type) {
	case string:
		str = t
	case *Message:
		out, err := json.Marshal(t)
		if err != nil {
			log.Error(err)
			return
		}
		str = string(out)

	}

	r.Lock()
	defer r.Unlock()
	for _, c := range r.clients {
		c.out <- str
	}
}

// Remove a client
func (r *Clients) removeClient(client *Client) {
	r.Lock()
	defer r.Unlock()

	for index, c := range r.clients {
		if c == client {
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
func (clients *Clients) WebsocketRoute(params martini.Params, receiver <-chan string, sender chan<- string, done <-chan bool, disconnect chan<- int, err <-chan error) (int, string) {
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
			fmt.Println("incoming message from webui on websocket", msg)
			clients.Router.Run(msg)

		case <-client.done:
			clients.removeClient(client)
			fmt.Println("waitgroup DONE")
			return 200, "OK"
		}
	}
}
