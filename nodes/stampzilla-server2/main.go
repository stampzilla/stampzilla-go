package main

import (
	"crypto/tls"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/contrib/ginrus"
	"github.com/gin-gonic/gin"
	"github.com/olahol/melody"
	"github.com/onrik/logrus/filename"
	"github.com/sirupsen/logrus"
	"github.com/soheilhy/cmux"
)

func main() {
	filenameHook := filename.NewHook()
	logrus.AddHook(filenameHook)

	r := gin.New()
	m := melody.New()

	r.Use(ginrus.Ginrus(logrus.StandardLogger(), time.RFC3339, true))

	// Startup the store
	store := NewStore()
	store.OnUpdate(broadcastNodeUpdate(m))

	// Setup melody
	m.Upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	m.HandleConnect(handleConnect(store))
	m.HandleMessage(handleMessage(m, store))
	m.HandleDisconnect(handleDisconnect(store))

	ca := &CA{}
	ca.LoadOrCreate()

	// Setup gin
	r.StaticFile("/", "./web/dist/index.html")
	r.StaticFile("/main.js", "./web/dist/main.js")
	r.Static("/assets", "./web/dist/assets")
	r.GET("/ca", handleDownloadCA(ca))

	r.GET("/ws", handleWs(m))

	cert, err := LoadOrCreate("localhost", ca)
	if err != nil {
		logrus.Fatal(err)
	}

	// Setup connection mux
	l, err := net.Listen("tcp", ":5000")
	if err != nil {
		logrus.Fatal(err)
	}

	mux := cmux.New(l)
	muxHTTP1 := mux.Match(cmux.HTTP1Fast())
	muxTLS := mux.Match(cmux.Any())

	// Start http server
	logger := log.New(logrus.StandardLogger().Writer(), "http: ", log.LstdFlags)
	http1 := &http.Server{
		Handler:      r,
		ReadTimeout:  time.Minute,
		WriteTimeout: time.Minute,
		ErrorLog:     logger,
	}
	go http1.Serve(muxHTTP1)

	// Start tls server
	tls := tls.NewListener(muxTLS, &tls.Config{
		Certificates: []tls.Certificate{*cert.TLS},
		ClientAuth:   tls.RequestClientCert,
	})
	go http1.Serve(tls)

	if err := mux.Serve(); !strings.Contains(err.Error(),
		"use of closed network connection") {
		logrus.Fatal(err)
	}
}

func handleWs(m *melody.Melody) func(c *gin.Context) {
	counter := 0
	return func(c *gin.Context) {
		counter++
		keys := make(map[string]interface{})
		keys["ID"] = strconv.Itoa(counter)
		r := c.Request

		if r.TLS != nil {
			certs := r.TLS.PeerCertificates
			logrus.Warn("HTTP CERTS", certs)
			keys["secure"] = true
		}

		// Accept the requested protocol
		// TODO: only accept known protocols
		if r.Header.Get("Sec-WebSocket-Protocol") != "" {
			c.Writer.Header().Set("Sec-WebSocket-Protocol", r.Header.Get("Sec-WebSocket-Protocol"))
			keys["protocol"] = r.Header.Get("Sec-WebSocket-Protocol")
		}

		m.HandleRequestWithKeys(c.Writer, c.Request, keys)
	}
}

func handleDownloadCA(ca *CA) func(c *gin.Context) {
	return func(c *gin.Context) {
		header := c.Writer.Header()
		header["Content-Type"] = []string{"application/x-x509-ca-cert"}
		//header["Content-Type"] = []string{"application/x-x509-user-cert"}
		//header["Content-Disposition"] = []string{"attachment; filename=stampzilla.crt"}

		file, err := os.Open("ca.crt")
		if err != nil {
			c.String(http.StatusOK, "%v", err)
			return
		}
		defer file.Close()

		io.Copy(c.Writer, file)
	}
}

func handleConnect(store *Store) func(s *melody.Session) {
	return func(s *melody.Session) {
		id, _ := s.Get("ID")
		t, exists := s.Get("protocol")

		store.AddOrUpdateConnection(id.(string), &Connection{
			Type:       t.(string),
			RemoteAddr: s.Request.RemoteAddr,
			Attributes: s.Keys,
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
