package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/olahol/melody"
	"github.com/onrik/logrus/filename"
	"github.com/sirupsen/logrus"
)

func main() {
	filenameHook := filename.NewHook()
	logrus.AddHook(filenameHook)

	r := gin.Default()
	m := melody.New()

	// Startup the store
	store := NewStore()
	store.OnUpdate(broadcastNodeUpdate(m))

	// Setup melody
	m.Upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	m.HandleConnect(handleConnect(store))
	m.HandleMessage(handleMessage(m, store))
	m.HandleDisconnect(handleDisconnect(store))

	// Setup gin
	r.StaticFile("/", "./web/dist/index.html")
	r.StaticFile("/main.js", "./web/dist/main.js")

	r.GET("/ws", handleWs(m))

	r.Run(":5000")
}

func handleWs(m *melody.Melody) func(c *gin.Context) {
	counter := 0
	return func(c *gin.Context) {
		counter += 1
		keys := make(map[string]interface{})
		keys["ID"] = strconv.Itoa(counter)

		// Accept the requested protocol
		// TODO: only accept known protocols
		if c.Request.Header.Get("Sec-WebSocket-Protocol") != "" {
			c.Writer.Header().Set("Sec-WebSocket-Protocol", c.Request.Header.Get("Sec-WebSocket-Protocol"))
			keys["protocol"] = c.Request.Header.Get("Sec-WebSocket-Protocol")
		}

		m.HandleRequestWithKeys(c.Writer, c.Request, keys)
	}
}

func handleConnect(store *Store) func(s *melody.Session) {
	return func(s *melody.Session) {
		id, _ := s.Get("ID")
		t, exists := s.Get("protocol")

		store.AddOrUpdateConnection(id.(string), &Connection{
			Type:       t.(string),
			RemoteAddr: s.Request.RemoteAddr,
		})

		if exists && t == "gui" {
			msg, err := NewMessage("nodes", store.GetNodes())
			if err != nil {
				logrus.Error(err)
				return
			}
			msg.Write(s)
		}
	}
}

func handleDisconnect(store *Store) func(s *melody.Session) {
	return func(s *melody.Session) {
		id, _ := s.Get("ID")
		store.RemoveConnection(id.(string))
	}
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
		case "update-node":
			handleNodeUpdate(m, s, store, data)
		}
	}
}

func handleNodeUpdate(m *melody.Melody, s *melody.Session, store *Store, data *Message) {
	node := &Node{}
	err := json.Unmarshal(data.Body, node)
	if err != nil {
		logrus.Error(err)
		return
	}

	store.AddOrUpdateNode(node)
}

func broadcastNodeUpdate(m *melody.Melody) func(store *Store) {
	return func(store *Store) {
		msg, err := NewMessage("nodes", store.GetNodes())
		if err != nil {
			logrus.Error(err)
			return
		}

		err = msg.WriteWithFilter(m, func(s *melody.Session) bool {
			v, exists := s.Get("protocol")
			return exists && v == "gui"
		})
		if err != nil {
			logrus.Error(err)
			return
		}

		msg, err = NewMessage("connections", store.GetConnections())
		if err != nil {
			logrus.Error(err)
			return
		}

		err = msg.WriteWithFilter(m, func(s *melody.Session) bool {
			v, exists := s.Get("protocol")
			return exists && v == "gui"
		})
		if err != nil {
			logrus.Error(err)
			return
		}
	}
}
