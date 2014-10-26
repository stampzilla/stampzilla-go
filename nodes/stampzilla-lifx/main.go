package main

import (
	//"github.com/bjeanes/go-lifx"

	"flag"
	"time"

	log "github.com/cihub/seelog"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/stamp/go-lifx/client"
	"github.com/stampzilla/stampzilla-go/nodes/basenode"
	"github.com/stampzilla/stampzilla-go/protocol"
)

// GLOBAL VARS
var node *protocol.Node
var state *State
var serverSendChannel chan interface{}
var serverRecvChannel chan protocol.Command
var lifxClient *client.Client

func init() { // {{{
	// Parse all commandline arguments, host and port parameters are added in the basenode init function
	flag.Parse()

	//Get a config with the correct parameters
	config := basenode.NewConfig()

	//Activate the config
	basenode.SetConfig(config)

	node = protocol.NewNode("lifx")

	//Create channels so we can communicate with the stampzilla-go server
	serverSendChannel = make(chan interface{})
	serverRecvChannel = make(chan protocol.Command)

	//Start communication with the server
	connectionState := basenode.Connect(serverSendChannel, serverRecvChannel)

	// This worker keeps track on our connection state, if we are connected or not
	go monitorState(connectionState, serverSendChannel)

	// This worker recives all incomming commands
	go serverRecv(serverRecvChannel)
} // }}}

// MAIN - This is run when the init function is done
func main() { /*{{{*/
	// Load logger
	logger, err := log.LoggerFromConfigAsFile("../logconfig.xml")
	if err != nil {
		panic(err)
	}
	log.ReplaceLogger(logger)

	// Create new node description

	state = NewState()
	node.SetState(state)

	lifxClient = client.New()
	go discoverWorker(lifxClient)

	select {}

} /*}}}*/

func discoverWorker(client *client.Client) {

	go monitorLampCollection(client.Lights)

	for {
		log.Info("Try to discover bulbs:")
		for _ = range client.Discover() {
		}

		<-time.After(60 * time.Second)
	}
}

// WORKER that monitors the list of lamps
func monitorLampCollection(lights *client.LightCollection) {
	for s := range lights.Lights.Changes {
		switch s.Event {
		case client.LampAdded:
			node.AddElement(&protocol.Element{
				Type: protocol.ElementTypeToggle,
				Name: s.Lamp.Label(),
				Command: &protocol.Command{
					Cmd:  "set",
					Args: []string{s.Lamp.Id()},
				},
				Feedback: "Devices[2].State",
			})
			node.AddElement(&protocol.Element{
				Type: protocol.ElementTypeColorPicker,
				Name: s.Lamp.Label(),
				Command: &protocol.Command{
					Cmd:  "color",
					Args: []string{s.Lamp.Id()},
				},
				Feedback: "Devices[4].State",
			})

			serverSendChannel <- node.Node()

			log.Warn("Added")
		case client.LampUpdated:
			log.Warn("Changed")
		case client.LampRemoved:
			log.Warn("Removed")
		}
	}
}

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

// WORKER that recives all incoming commands
func serverRecv(recv chan protocol.Command) {
	for d := range recv {
		processCommand(d)
	}
}

// This is called on each incoming command
func processCommand(cmd protocol.Command) {
	log.Info("Incoming command from server:", cmd)

	lamp, err := lifxClient.Lights.GetById(cmd.Args[0])
	if err != nil {
		log.Error(err)
		return
	}

	switch cmd.Cmd {
	case "set":
		switch cmd.Params[0] {
		case "true":
			lamp.TurnOn()
		case "false":
			lamp.TurnOff()
		}
	case "color":
		c, err := colorful.Hex(cmd.Params[0])
		if err != nil {
			log.Error(err)
			return
		}

		h, s, v := c.Hsv()

		lamp.SetState(h, s, v)
	}
}
