package webserver

import (
	"crypto/tls"
	"time"

	"github.com/jonaz/gograce"
	"github.com/olahol/melody"
	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/handlers"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/models"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/store"
)

type Secure struct {
	*Webserver
}

func NewSecure(s *store.Store, conf *models.Config, wsh handlers.WebsocketHandler, m *melody.Melody) *Secure {
	return &Secure{
		Webserver: New(s, conf, wsh, m),
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
