package main

import (
	"encoding/json"
	"fmt"
	"sync"

	log "github.com/cihub/seelog"
	"github.com/go-martini/martini"
	"github.com/gorilla/websocket"
)

var clients *Clients

type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
	To   string      `json:"to"`
}
type Clients struct {
	sync.Mutex
	clients []*Client
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
	clients.messageOtherClients(&Message{Type: "all", Data: nodes.All()})
}

// Message all the other clients
func (r *Clients) messageOtherClients(msg *Message) {
	r.Lock()
	defer r.Unlock()
	for _, c := range r.clients {
		c.out <- msg
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

// Remove a client
func (r *Clients) disconnectAll() {
	r.Lock()
	defer r.Unlock()

	for _, c := range r.clients {
		c.disconnect <- websocket.CloseGoingAway
	}
}

func newClients() *Clients {
	return &Clients{sync.Mutex{}, make([]*Client, 0)}
}
func websocketRoute(params martini.Params, receiver <-chan *Message, sender chan<- *Message, done <-chan bool, disconnect chan<- int, err <-chan error) (int, string) {
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
			node := nodes.Search(msg.To)
			if node != nil {
				jsonToSend, err := json.Marshal(&msg.Data)
				if err != nil {
					log.Error(err)
					continue
				}
				node.conn.Write(jsonToSend)
			}
		case <-client.done:
			clients.removeClient(client)
			fmt.Println("waitgroup DONE")
			return 200, "OK"
		}
	}
}
