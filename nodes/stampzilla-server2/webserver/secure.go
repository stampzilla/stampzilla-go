package webserver

import (
	"crypto/tls"
	"encoding/json"
	"time"

	"github.com/jonaz/gograce"
	"github.com/olahol/melody"
	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/models"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/store"
)

type Secure struct {
	*Webserver
}

func NewSecure(s *store.Store, conf *models.Config) *Secure {
	return &Secure{
		Webserver: New(s, conf),
	}

}

func (ws *Secure) Start(addr string, tlsConfig *tls.Config) chan struct{} {

	server, done := gograce.NewServerWithTimeout(10 * time.Second)

	server.Handler = ws.Init()
	server.Addr = addr
	server.TLSConfig = tlsConfig

	go func() {
		logrus.Error(server.ListenAndServeTLS("", ""))
	}()
	return done
}

func (ws *Secure) handleConnect(store *store.Store) func(s *melody.Session) {
	return func(s *melody.Session) {
		id, _ := s.Get("ID")
		t, exists := s.Get("protocol")

		store.AddOrUpdateConnection(id.(string), &models.Connection{
			Type:       t.(string),
			RemoteAddr: s.Request.RemoteAddr,
			Attributes: s.Keys,
		})

		if exists && t == "gui" {
			msg, err := models.NewMessage("nodes", store.GetNodes())
			if err != nil {
				logrus.Error(err)
				return
			}
			msg.Write(s)
		}
	}
}

func (ws *Secure) handleMessage(m *melody.Melody, store *store.Store) func(s *melody.Session, msg []byte) {
	return func(s *melody.Session, msg []byte) {
		data, err := models.ParseMessage(msg)
		if err != nil {
			logrus.Error(err)
			return
		}

		switch data.Type {
		case "update-node":
			ws.handleNodeUpdate(m, s, store, data)
		}
	}
}

func (ws *Secure) handleNodeUpdate(m *melody.Melody, s *melody.Session, store *store.Store, data *models.Message) {
	node := &models.Node{}
	err := json.Unmarshal(data.Body, node)
	if err != nil {
		logrus.Error(err)
		return
	}

	store.AddOrUpdateNode(node)
}
