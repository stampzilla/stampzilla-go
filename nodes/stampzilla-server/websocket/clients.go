package websocket

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	log "github.com/cihub/seelog"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/pborman/uuid"
)

const (
	// Time allowed to write the file to the client.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the client.
	pongWait = 60 * time.Second

	// Send pings to client with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
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
	out        chan<- *Message
	done       <-chan bool
	err        <-chan error
	Id         string
	disconnect chan int
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
	clientsToRemove := make([]*Client, 0)

	for _, c := range r.clients {
		select {
		case c.out <- msg:
			// Everything went well :)
		case <-time.After(time.Second):
			log.Warn("Failed writing to websocket: timeout (", c.Id, ")")
			clientsToRemove = append(clientsToRemove, c)
		}
	}

	r.Unlock()
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

func (clients *Clients) WebsocketRoute(c *gin.Context) {
	conn, err := websocket.Upgrade(c.Writer, c.Request, nil, 1024, 1024)
	if err != nil {
		http.Error(c.Writer, "Websocket error", 400)
		log.Error(err)
		return
	}

	out := make(chan *Message)
	done := make(chan bool)
	wsErr := make(chan error)
	disconnect := make(chan int)

	defer func() {
		close(out)
		close(done)
		close(wsErr)
	}()

	client := &Client{out, done, wsErr, uuid.New(), disconnect}

	go senderWorker(conn, out)
	go disconnectWorker(conn, client)

	clients.appendClient(client)

	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		//log.Debug("Got pong response from browser")
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	//Listen to websocket
	for {
		msg := &Message{}
		err := conn.ReadJSON(msg)
		if err != nil {
			log.Info(conn.LocalAddr(), " Disconnected")
			clients.removeClient(client)
			return
		}
		go clients.Router.Run(msg)
	}
}

func disconnectWorker(conn *websocket.Conn, c *Client) {
	defer close(c.disconnect)
	for code := range c.disconnect {
		log.Debug("Closing websocket")
		conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(code, ""), time.Now().Add(writeWait))
		if err := conn.Close(); err != nil {
			log.Error("Connection could not be closed: ", err)
		}
		return
	}
}
func senderWorker(conn *websocket.Conn, out chan *Message) {
	pingTicker := time.NewTicker(pingPeriod)
	defer func() {
		pingTicker.Stop()
	}()
	for {
		select {
		case msg, opened := <-out:
			if !opened {
				log.Debug("websocket: Sendchannel closed stopping pingTicket and senderWorker")
				return
			}
			if err := conn.WriteJSON(msg); err != nil {
				log.Error(err)
			}
		case <-pingTicker.C:
			if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(writeWait)); err != nil {
				log.Error(err)
				return
			}

		}
	}
}
