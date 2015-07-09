package main

import (
	"flag"

	log "github.com/cihub/seelog"
	"github.com/stampzilla/stampzilla-go/nodes/basenode"
	"github.com/stampzilla/stampzilla-go/protocol"
)

// GLOBAL VARS
var node *protocol.Node

type TargetState struct {
	Targets    map[string]*Target
	connection *basenode.Connection
}

var state TargetState

func (s TargetState) GetState() interface{} {
	return s
}

func (s *TargetState) Add(t *Target) {
	s.Targets[t.Name] = t

	node.AddElement(&protocol.Element{
		Type:     protocol.ElementTypeText,
		Name:     t.Name,
		Feedback: `Targets["` + t.Name + `"].Online`,
	})

	node.AddElement(&protocol.Element{
		Type:     protocol.ElementTypeText,
		Name:     t.Name,
		Feedback: `Targets["` + t.Name + `"].Lag`,
	})

	t.start(s.connection)
}

// MAIN - This is run when the init function is done
func main() { /*{{{*/
	log.Info("Starting PINGER node")
	// Create new node description
	// Parse all commandline arguments, host and port parameters are added in the basenode init function
	flag.Parse()

	//Get a config with the correct parameters
	config := basenode.NewConfig()

	//Activate the config
	basenode.SetConfig(config)

	node = protocol.NewNode("pinger")

	//Start communication with the server
	connection := basenode.Connect()

	// Thit worker keeps track on our connection state, if we are connected or not
	go monitorState(connection)

	// This worker recives all incomming commands
	go serverRecv(connection.Receive)

	state = TargetState{
		Targets:    make(map[string]*Target),
		connection: connection,
	}
	node.SetState(state)

	t := &Target{
		Name: "stamps",
		Ip:   "10.21.10.115",
	}

	state.Add(t)

	select {}
} /*}}}*/

// WORKER that monitors the current connection state
func monitorState(connection *basenode.Connection) {
	for s := range connection.State {
		switch s {
		case basenode.ConnectionStateConnected:
			connection.Send <- node.Node()
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
