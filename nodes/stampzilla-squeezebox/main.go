package main

import (
	"flag"

	log "github.com/cihub/seelog"
	"github.com/stampzilla/stampzilla-go/nodes/basenode"
	"github.com/stampzilla/stampzilla-go/pkg/notifier"
	"github.com/stampzilla/stampzilla-go/protocol"
)

var VERSION string = "dev"
var BUILD_DATE string = ""

var notify *notifier.Notify

type Config struct {
	Host     string `json:"host"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func main() {
	log.Info("Starting squeezebox node")

	// Parse all commandline arguments, host and port parameters are added in the basenode init function
	flag.Parse()

	//Get a config with the correct parameters
	config := basenode.NewConfig()

	//Activate the config
	basenode.SetConfig(config)

	nc := &Config{}
	err := config.NodeSpecific(&nc)
	if err != nil {
		log.Error(err)
		return
	}

	node := protocol.NewNode("squeezebox")
	node.Version = VERSION
	node.BuildDate = BUILD_DATE

	//Start communication with the server
	connection := basenode.Connect()
	notify = notifier.New(connection)
	notify.SetSource(node)

	// Thit worker keeps track on our connection state, if we are connected or not
	go monitorState(node, connection)

	state := NewState()
	node.SetState(state)

	//state.AddDevice("1", "Dev1", false)
	//state.AddDevice("2", "Dev2", true)

	// This worker recives all incomming commands

	s := NewSqueezebox()
	processor := NewProcessor(node, connection, state, s)
	go serverRecv(processor)
	go squeezeboxReader(s, processor)
	err = s.Connect(nc.Host, nc.Username, nc.Password)
	if err != nil {
		log.Error(err)
		return
	}

	s.Send("subscribe power,alarm,pause,play,stop,client,mixer,playlist")
	s.Send("players 0 999")

	select {}
}

func squeezeboxReader(s *squeezebox, processor *Processor) {
	for v := range s.Read() {
		log.Info(v)
		processor.ProcessSqueezeboxCommand(v)

	}

}

// WORKER that monitors the current connection state
func monitorState(node *protocol.Node, connection basenode.Connection) {
	for s := range connection.State() {
		switch s {
		case basenode.ConnectionStateConnected:
			connection.Send(node.Node())
		case basenode.ConnectionStateDisconnected:
		}
	}
}

// WORKER that recives all incomming commands
func serverRecv(processor *Processor) {
	for cmd := range processor.connection.Receive() {
		processor.ProcessServerCommand(cmd)
	}
}
