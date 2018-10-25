package main

import (
	"encoding/json"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/olahol/melody"
)

func main() {

	r := gin.Default()
	m := melody.New()

	store := NewStore()

	r.StaticFile("/", "./web/dist/index.html")
	r.StaticFile("/main.js", "./web/dist/main.js")

	r.GET("/ws", func(c *gin.Context) {
		m.HandleRequest(c.Writer, c.Request)
	})

	m.Upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	m.HandleMessage(handleMessage(m, store))

	r.Run(":5000")
}

func handleMessage(m *melody.Melody, store *Store) func(s *melody.Session, msg []byte) {
	return func(s *melody.Session, msg []byte) {
		data := &Message{}
		err := json.Unmarshal(msg, data)
		if err != nil {
			logrus.Error(err)
			return
		}

		switch data.Type {
		case "all-nodes":
			handleAllNodes(s, store, msg)
		case "update-node":
			handleNodeUpdate(m, s, store, msg, data)
		}
	}
}

func handleAllNodes(s *melody.Session, store *Store, msg []byte) {
	s.Set("all-nodes", true)
	WriteJSON(s, "nodes", store.GetNodes())
}

func handleNodeUpdate(m *melody.Melody, s *melody.Session, store *Store, msg []byte, data *Message) {
	node := &Node{}
	err := json.Unmarshal(data.Body, node)
	if err != nil {
		logrus.Error(err)
		return
	}

	store.AddOrUpdateNode(node)

	msg, err = NewMessageJSON("nodes", store.GetNodes())
	if err != nil {
		logrus.Error(err)
		return
	}
	m.BroadcastFilter(msg, func(s *melody.Session) bool {
		v, exists := s.Get("all-nodes")
		return exists && v == true
	})
}

func NewMessageJSON(t string, body interface{}) ([]byte, error) {
	b, err := json.Marshal(body)

	if err != nil {
		return nil, err
	}

	msg, err := json.Marshal(&Message{
		Type: t,
		Body: json.RawMessage(b),
	})

	return msg, err
}

func WriteJSON(s *melody.Session, t string, body interface{}) error {

	msg, err := NewMessageJSON(t, body)
	if err != nil {
		return err
	}

	return s.Write(msg)
}

type Message struct {
	Type string          `json:"type"`
	Body json.RawMessage `json:"body"`
}
