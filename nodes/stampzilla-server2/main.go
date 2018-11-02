package main

import (
	"crypto/tls"

	"github.com/olahol/melody"
	"github.com/onrik/logrus/filename"
	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/models"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/store"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/webserver"
)

func main() {

	config := &models.Config{}
	config.MustLoad()

	filenameHook := filename.NewHook()
	logrus.AddHook(filenameHook)

	store := store.New()
	httpServer := webserver.New(store, config)
	tlsServer := webserver.NewSecure(store, config)

	// Startup the store
	ca := &CA{}
	ca.LoadOrCreate()
	cert, err := LoadOrCreate("server", ca)
	if err != nil {
		logrus.Fatal(err)
	}

	config.WriteToFile("config.json")

	done := httpServer.Start(":8080")
	tlsDone := tlsServer.Start(":6443", &tls.Config{
		Certificates: []tls.Certificate{*cert.TLS},
		ClientAuth:   tls.RequestClientCert,
	})

	//store.OnUpdate(broadcastNodeUpdate(httpServer.Melody))
	store.OnUpdate(broadcastNodeUpdate(tlsServer.Melody))

	<-done
	<-tlsDone
	config.WriteToFile("config.json")
}

func broadcastNodeUpdate(m *melody.Melody) func(*store.Store) error {
	return func(store *store.Store) error {

		msg, err := models.NewMessage("nodes", store.GetNodes())
		if err != nil {
			return err
		}

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

		return msg.WriteWithFilter(m, func(s *melody.Session) bool {
			v, exists := s.Get("protocol")
			return exists && v == "gui"
		})
	}
}
