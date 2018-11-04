package main

import (
	"crypto/tls"

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

func main() {

	config := &models.Config{}
	config.MustLoad()

	filenameHook := filename.NewHook()
	logrus.AddHook(filenameHook)

	// Startup the store
	ca, err := ca.LoadOrCreate("ca")
	if err != nil {
		logrus.Fatal(err)
	}
	err = ca.LoadOrCreate("server")
	if err != nil {
		logrus.Fatal(err)
	}

	//logrus.Infof("tls: %#v", ca.CATLS)
	//logrus.Infof("x509: %#v", ca.CAX509)
	//logrus.Info("x509 subject:", ca.CAX509.Subject)

	insecureMelody := melody.New()
	//TODO i dont like melody anymore.. raw gorilla seems fine?
	insecureMelody.Config.MaxMessageSize = 0
	secureMelody := melody.New()
	//TODO i dont like melody anymore.. raw gorilla seems fine?
	secureMelody.Config.MaxMessageSize = 0

	insecureSender := websocket.NewWebsocketSender(insecureMelody)
	secureSender := websocket.NewWebsocketSender(secureMelody)

	store := store.New()
	httpServer := webserver.New(store, config, handlers.NewInSecureWebsockerHandler(store, config, insecureSender, ca), insecureMelody)
	tlsServer := webserver.NewSecure(store, config, handlers.NewSecureWebsockerHandler(store, config, secureSender), secureMelody)

	config.Save("config.json")

	done := httpServer.Start(":8080")
	tlsDone := tlsServer.Start(":6443", &tls.Config{
		Certificates: []tls.Certificate{*ca.TLS},
		ClientAuth:   tls.RequestClientCert,
	})

	//store.OnUpdate(broadcastNodeUpdate(httpServer.Melody))
	store.OnUpdate(broadcastNodeUpdate(tlsServer.Melody))

	<-done
	<-tlsDone
	config.Save("config.json")
}

func broadcastNodeUpdate(m *melody.Melody) func(*store.Store) error {
	return func(store *store.Store) error {

		msg, err := models.NewMessage("nodes", store.GetNodes())
		if err != nil {
			return err
		}

		// TODO move this to websocket.Sender and depend on it here. Does not belong on *Message
		err = msg.WriteWithFilter(m, func(s *melody.Session) bool {
			v, exists := s.Get("protocol")
			return exists && v == "gui"
		})
		if err != nil {
			return err
		}

		msg, err = models.NewMessage("connections", store.GetConnections())
		if err != nil {
			return err
		}

		// TODO move this to websocket.Sender and depend on it here. Does not belong on *Message
		return msg.WriteWithFilter(m, func(s *melody.Session) bool {
			v, exists := s.Get("protocol")
			return exists && v == "gui"
		})
	}
}
