package servermain

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"

	"github.com/olahol/melody"
	"github.com/onrik/logrus/filename"
	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/ca"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/handlers"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/models"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/store"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/webserver"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/websocket"
)

// Main contains deps used by server that will be exposed so we can write good end to end tests
type Main struct {
	Config     *models.Config
	Store      *store.Store
	HTTPServer *webserver.Webserver
	TLSServer  *webserver.Webserver
	CA         *ca.CA
}

// New creates a new main
func New(config *models.Config, store *store.Store) *Main {
	return &Main{
		Config: config,
		Store:  store,
	}
}

func (c *Main) Run() {
	done := c.HTTPServer.Start(":"+c.Config.Port, nil)
	tlsDone := c.TLSServer.Start(":"+c.Config.TLSPort, c.TLSConfig())
	<-done
	<-tlsDone
	c.Config.Save("config.json")
}

func (c *Main) TLSConfig() *tls.Config {
	caCert, err := ioutil.ReadFile("ca.crt")
	if err != nil {
		log.Fatal(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	return &tls.Config{
		// Needed to verify client certificates
		ClientCAs:    caCertPool,
		Certificates: []tls.Certificate{*c.CA.TLS},
		ClientAuth:   tls.VerifyClientCertIfGiven,
	}
}

// Init initializes the web handlers. Could be used to start server or to test it
func (m *Main) Init() {
	var err error

	filenameHook := filename.NewHook()
	logrus.AddHook(filenameHook)

	m.CA, err = ca.LoadOrCreate()
	if err != nil {
		logrus.Fatal(err)
	}
	err = m.CA.LoadOrCreate("localhost")
	if err != nil {
		logrus.Fatal(err)
	}

	insecureMelody := melody.New()
	//TODO i dont like melody anymore.. raw gorilla seems fine?
	insecureMelody.Config.MaxMessageSize = 0
	secureMelody := melody.New()
	//TODO i dont like melody anymore.. raw gorilla seems fine?
	secureMelody.Config.MaxMessageSize = 0

	insecureSender := websocket.NewWebsocketSender(insecureMelody)
	secureSender := websocket.NewWebsocketSender(secureMelody)

	if err = m.Store.LoadFromDisk(); err != nil {
		log.Fatalf("Failed to load state from disk: %s", err)
	}
	m.HTTPServer = webserver.New(m.Store, m.Config, handlers.NewInSecureWebsockerHandler(m.Store, m.Config, insecureSender, m.CA), insecureMelody)
	m.TLSServer = webserver.New(m.Store, m.Config, handlers.NewSecureWebsockerHandler(m.Store, m.Config, secureSender), secureMelody)

	m.Config.Save("config.json")

	m.Store.OnUpdate("nodes", broadcastNodeUpdate(secureSender))
	m.Store.OnUpdate("connections", broadcastConnectionsUpdate(secureSender))
	m.Store.OnUpdate("devices", broadcastDevicesUpdate(secureSender))

}

func broadcastNodeUpdate(sender websocket.Sender) func(*store.Store) error {
	return func(store *store.Store) error {
		return sender.SendToProtocol("gui", "nodes", store.GetNodes())
	}
}

func broadcastConnectionsUpdate(sender websocket.Sender) func(*store.Store) error {
	return func(store *store.Store) error {
		return sender.SendToProtocol("gui", "connections", store.GetConnections())
	}
}

func broadcastDevicesUpdate(sender websocket.Sender) func(*store.Store) error {
	return func(store *store.Store) error {
		return sender.SendToProtocol("gui", "devices", store.GetDevices())
	}
}
