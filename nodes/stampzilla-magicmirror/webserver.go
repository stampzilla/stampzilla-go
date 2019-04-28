package main

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/olahol/melody"
	"github.com/rakyll/statik/fs"
	"github.com/sirupsen/logrus"

	_ "github.com/stampzilla/stampzilla-go/nodes/stampzilla-magicmirror/statik"
)

type Message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

var m = melody.New()
var lastMessages = make(map[string][]byte)

func initWebserver() {
	r := gin.Default()

	statikFS, err := fs.New()
	if err != nil {
		logrus.Warn("Statik FS failed:", err)
	}
	if err == nil { // we only service GUI if statik files can be found
		r.GET("/manifest.json", gin.WrapH(http.FileServer(statikFS)))
		r.NoRoute(func(c *gin.Context) {
			cspMiddleware()(c)

			// Check if the file exists
			_, err := statikFS.Open(c.Request.URL.Path)
			if err != nil {
				c.Request.URL.Path = "/" // force us to always return index.html and not the requested page to be compatible with HTML5 routing
			}

			http.FileServer(statikFS).ServeHTTP(c.Writer, c.Request)
		})

		//r.StaticFile("/", "./web/build/index.html")
		//r.StaticFile("/manifest.json", "./web/build/manifest.json")
		//r.Static("/static", "./web/build/static")
	}

	r.GET("/ws", func(c *gin.Context) {
		m.HandleRequest(c.Writer, c.Request)
	})

	r.GET("/proxy", func(c *gin.Context) {
		response, err := http.Get(c.Query("url"))
		if err != nil || response.StatusCode != http.StatusOK {
			c.Status(http.StatusServiceUnavailable)
			return
		}

		reader := response.Body
		contentLength := response.ContentLength
		contentType := response.Header.Get("Content-Type")
		extraHeaders := map[string]string{}

		c.DataFromReader(http.StatusOK, contentLength, contentType, reader, extraHeaders)
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

func cspMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Security-Policy", "worker-src 'self';")
		c.Next()
	}
}
