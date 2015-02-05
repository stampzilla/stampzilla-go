package main

import (
	"flag"

	log "github.com/cihub/seelog"
	"github.com/stampzilla/stampzilla-go/nodes/basenode"
	"github.com/stampzilla/stampzilla-go/protocol"
)

// MAIN - This is run when the init function is done
func main() { /*{{{*/
	log.Info("Starting SIMPLE node")

	// Parse all commandline arguments, host and port parameters are added in the basenode init function
	flag.Parse()

	//Get a config with the correct parameters
	config := basenode.NewConfig()

	//Activate the config
	basenode.SetConfig(config)

	node := protocol.NewNode("simple")

	//Start communication with the server
	connection := basenode.Connect()

	// Thit worker keeps track on our connection state, if we are connected or not
	go monitorState(node, connection)

	// This worker recives all incomming commands
	go serverRecv(connection.Receive)

	// Describe available actions
	node.AddAction("set", "Set", []string{"Devices.Id"})
	node.AddAction("toggle", "Toggle", []string{"Devices.Id"})
	node.AddAction("dim", "Dim", []string{"Devices.Id", "value"})

	node.AddElement(&protocol.Element{
		Type: protocol.ElementTypeText,
		Name: "Example text",
		Command: &protocol.Command{
			Cmd:  "set",
			Args: []string{"1"},
		},
		Feedback: "Devices[0].State",
	})
	node.AddElement(&protocol.Element{
		Type: protocol.ElementTypeButton,
		Name: "Example button",
		Command: &protocol.Command{
			Cmd:  "set",
			Args: []string{"1"},
		},
		Feedback: "Devices[1].State",
	})
	node.AddElement(&protocol.Element{
		Type: protocol.ElementTypeToggle,
		Name: "Example toggle",
		Command: &protocol.Command{
			Cmd:  "set",
			Args: []string{"1"},
		},
		Feedback: "Devices[2].State",
	})
	node.AddElement(&protocol.Element{
		Type: protocol.ElementTypeSlider,
		Name: "Example slider",
		Command: &protocol.Command{
			Cmd:  "set",
			Args: []string{"1"},
		},
		Feedback: "Devices[3].State",
	})
	node.AddElement(&protocol.Element{
		Type: protocol.ElementTypeColorPicker,
		Name: "Example color picker",
		Command: &protocol.Command{
			Cmd:  "set",
			Args: []string{"1"},
		},
		Feedback: "Devices[4].State",
	})

	state := NewState()
	node.SetState(state)

	state.AddDevice("1", "Dev1", "OFF")
	state.AddDevice("2", "Dev2", "ON")

	select {}
} /*}}}*/

// WORKER that monitors the current connection state
func monitorState(node *protocol.Node, connection *basenode.Connection) {
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
