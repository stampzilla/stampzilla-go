package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"os"

	"github.com/koding/multiconfig"
	"github.com/olahol/melody"
	"github.com/onrik/logrus/filename"
	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/models"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/store"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/webserver"
)

func main() {

	config := &models.Config{}
	m := loadMultiConfig()
	m.MustLoad(config)

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

	saveConfigToFile(config)

	done := httpServer.Start(":8080")
	tlsDone := tlsServer.Start(":6443", &tls.Config{
		Certificates: []tls.Certificate{*cert.TLS},
		ClientAuth:   tls.RequestClientCert,
	})

	//store.OnUpdate(broadcastNodeUpdate(httpServer.Melody))
	store.OnUpdate(broadcastNodeUpdate(tlsServer.Melody))

	<-done
	<-tlsDone
	saveConfigToFile(config)
}

func saveConfigToFile(config *models.Config) {
	configFile, err := os.Create("config.json")
	if err != nil {
		logrus.Error("creating config file", err.Error())
	}

	logrus.Info("Save config: ", config)
	var out bytes.Buffer
	b, err := json.MarshalIndent(config, "", "\t")
	if err != nil {
		logrus.Error("error marshal json", err)
	}
	json.Indent(&out, b, "", "\t")
	out.WriteTo(configFile)
}

func loadMultiConfig() *multiconfig.DefaultLoader {
	loaders := []multiconfig.Loader{}

	// Read default values defined via tag fields "default"
	loaders = append(loaders, &multiconfig.TagLoader{})

	if _, err := os.Stat("config.json"); err == nil {
		loaders = append(loaders, &multiconfig.JSONLoader{Path: "config.json"})
	}

	e := &multiconfig.EnvironmentLoader{}
	e.Prefix = "STAMPZILLA"
	f := &multiconfig.FlagLoader{}
	f.EnvPrefix = "STAMPZILLA"

	loaders = append(loaders, e, f)
	loader := multiconfig.MultiLoader(loaders...)

	d := &multiconfig.DefaultLoader{}
	d.Loader = loader
	d.Validator = multiconfig.MultiValidator(&multiconfig.RequiredValidator{})
	return d

}

func broadcastNodeUpdate(m *melody.Melody) func(*store.Store) {
	return func(store *store.Store) {

		msg, err := models.NewMessage("nodes", store.GetNodes())
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

		msg, err = models.NewMessage("connections", store.GetConnections())
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
