package servermain

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"

	"github.com/olahol/melody"
	"github.com/onrik/logrus/filename"
	"github.com/sirupsen/logrus"
	"github.com/stamp/mdns"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/ca"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/handlers"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/logic"
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
func New(config *models.Config) *Main {
	return &Main{
		Config: config,
	}
}

func (c *Main) Run() {
	done := c.HTTPServer.Start(":"+c.Config.Port, nil)
	tlsDone := c.TLSServer.Start(":"+c.Config.TLSPort, c.TLSConfig())

	// Setup and start mDNS
	if port, err := strconv.Atoi(c.Config.Port); err == nil {
		host, _ := os.Hostname()
		info := []string{"stampzilla-go"}
		mdnsService, _ := mdns.NewMDNSService(host, "_stampzilla._tcp", "", "", port, nil, info)
		mdnsServer, _ := mdns.NewServer(&mdns.Config{Zone: mdnsService})
		defer mdnsServer.Shutdown()
	}

	// start logic

	ctx, cancel := context.WithCancel(context.Background())
	c.Store.Logic.Start(ctx)
	c.Store.Scheduler.Start(ctx)

	<-done
	<-tlsDone
	cancel() // stop logic
	c.Store.Logic.Wait()
	c.Config.Save("config.json")
}

func (c *Main) TLSConfig() *tls.Config {
	caCert, err := ioutil.ReadFile(path.Join("certificates", "ca.crt"))
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

	if m.Config.LogLevel != "" {
		lvl, err := logrus.ParseLevel(m.Config.LogLevel)
		if err != nil {
			logrus.Fatal(err)
			return
		}
		logrus.SetLevel(lvl)
	}

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

	sss := logic.NewSavedStateStore()
	l := logic.New(sss, secureSender)
	scheduler := logic.NewScheduler(sss, secureSender)
	m.Store = store.New(l, scheduler, sss)
	m.CA.SetStore(m.Store)

	if err = m.Store.Load(); err != nil {
		log.Fatalf("Failed to load state from disk: %s", err)
	}
	m.HTTPServer = webserver.New(
		m.Store,
		m.Config,
		handlers.NewInSecureWebsockerHandler(m.Store, m.Config, insecureSender, m.CA),
		insecureMelody,
	)
	m.TLSServer = webserver.New(
		m.Store,
		m.Config,
		handlers.NewSecureWebsockerHandler(m.Store, m.Config, secureSender, m.CA),
		secureMelody,
	)

	m.Config.Save("config.json")

	m.Store.OnUpdate(handlers.BroadcastUpdate(secureSender))
}
