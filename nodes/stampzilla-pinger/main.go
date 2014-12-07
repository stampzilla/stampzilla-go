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
	Targets map[string]*Target
}

var state TargetState
var serverSendChannel chan interface{}
var serverRecvChannel chan protocol.Command

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
}

// INIT - The first function to run
func init() { // {{{
	// Parse all commandline arguments, host and port parameters are added in the basenode init function
	flag.Parse()

	//Get a config with the correct parameters
	config := basenode.NewConfig()

	//Activate the config
	basenode.SetConfig(config)

	node = protocol.NewNode("pinger")

	//Create channels so we can communicate with the stampzilla-go server
	serverSendChannel = make(chan interface{})
	serverRecvChannel = make(chan protocol.Command)

	//Start communication with the server
	connectionState := basenode.Connect(serverSendChannel, serverRecvChannel)

	// Thit worker keeps track on our connection state, if we are connected or not
	go monitorState(connectionState, serverSendChannel)

	// This worker recives all incomming commands
	go serverRecv(serverRecvChannel)
} // }}}

// MAIN - This is run when the init function is done
func main() { /*{{{*/
	log.Info("Starting PINGER node")
	// Create new node description

	state = TargetState{
		Targets: make(map[string]*Target),
	}
	node.SetState(state)

	t := &Target{
		Name: "stamps",
		Ip:   "10.21.10.148",
	}

	t.start()

	state.Add(t)

	select {}
} /*}}}*/

// WORKER that monitors the current connection state
func monitorState(connectionState chan int, send chan interface{}) {
	for s := range connectionState {
		switch s {
		case basenode.ConnectionStateConnected:
			send <- node.Node()
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
