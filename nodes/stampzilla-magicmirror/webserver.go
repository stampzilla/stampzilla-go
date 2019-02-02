package main

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/olahol/melody"
)

type Message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

var m = melody.New()
var lastMessages = make(map[string][]byte)

func initWebserver() {
	r := gin.Default()

	r.StaticFile("/", "./web/build/index.html")
	r.StaticFile("/manifest.json", "./web/build/manifest.json")
	r.Static("/static", "./web/build/static")

	r.GET("/ws", func(c *gin.Context) {
		m.HandleRequest(c.Writer, c.Request)
	})

	m.HandleConnect(func(s *melody.Session) {
		for _, message := range lastMessages {
			s.Write(message)
		}
	})

	r.Run(":8089")
}

func forwardAs(t string) func(data json.RawMessage) error {
	return func(data json.RawMessage) error {
		msg := &Message{
			Type: t,
			Data: data,
		}

		b, err := json.Marshal(msg)
		if err != nil {
			return err
		}

		lastMessages[t] = b

		m.Broadcast(b)
		return nil
	}
}
