package main

import (
	"flag"
	"fmt"
	"time"

	log "github.com/cihub/seelog"
	"github.com/stampzilla/stampzilla-go/nodes/basenode"
	"github.com/stampzilla/stampzilla-go/protocol"

	"github.com/stampzilla/gocast/discovery"
)

var state State

func main() { /*{{{*/
	log.Info("Starting CHROMECAST node")

	// Parse all commandline arguments, host and port parameters are added in the basenode init function
	flag.Parse()

	//Get a config with the correct parameters
	config := basenode.NewConfig()

	//Activate the config
	basenode.SetConfig(config)

	node := protocol.NewNode("chromecast")

	//Start communication with the server
	connection := basenode.Connect()

	// Thit worker keeps track on our connection state, if we are connected or not
	go monitorState(node, connection)

	// This worker recives all incomming commands
	go serverRecv(connection.Receive())

	state = State{
		connection: &connection,
		node:       node,
	}
	node.SetState(&state.Devices)

	discovery := discovery.NewService()

	go discoveryListner(discovery)
	discovery.Periodic(time.Second * 10)

	select {}
} /*}}}*/

func discoveryListner(discovery *discovery.Service) {
	for device := range discovery.Found() {
		fmt.Printf("New device discoverd: %#v \n", device)

		NewChromecast(device)
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
func serverRecv(recv chan protocol.Command) {
	for d := range recv {
		processCommand(d)
	}
}

// THis is called on each incomming command
func processCommand(cmd protocol.Command) {
	log.Info("Incoming command from server:", cmd)
}
