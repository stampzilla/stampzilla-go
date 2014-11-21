package main

import (
	"flag"

	log "github.com/cihub/seelog"
	"github.com/stampzilla/stampzilla-go/nodes/basenode"
	"github.com/stampzilla/stampzilla-go/protocol"
)

// MAIN - This is run when the init function is done
func main() { /*{{{*/
	// Load logger
	logger, err := log.LoggerFromConfigAsFile("../logconfig.xml")
	if err != nil {
		panic(err)
	}
	log.ReplaceLogger(logger)

	log.Info("Starting CHROMECAST node")

	// Parse all commandline arguments, host and port parameters are added in the basenode init function
	flag.Parse()

	//Get a config with the correct parameters
	config := basenode.NewConfig()

	//Activate the config
	basenode.SetConfig(config)

	node := protocol.NewNode("chromecast")

	//Create channels so we can communicate with the stampzilla-go server
	serverSendChannel := make(chan interface{})
	serverRecvChannel := make(chan protocol.Command)

	//Start communication with the server
	connectionState := basenode.Connect(serverSendChannel, serverRecvChannel)

	// Thit worker keeps track on our connection state, if we are connected or not
	go monitorState(node, connectionState, serverSendChannel)

	// This worker recives all incomming commands
	go serverRecv(serverRecvChannel)

	//Start chromecast monitoring
	chromecast := NewChromecast()
	node.SetState(chromecast.Devices)

	go func() {
		for {
			event := <-chromecast.Events
			if chromecast.Events == nil {
				return
			}

			log.Warn("EVENT: ", event.Name)
			switch event.Name {
			case "Added":
				dev := chromecast.Devices.Get(event.Args[0])

				node.AddElement(&protocol.Element{
					Type: protocol.ElementTypeText,
					Name: dev.Name,
					Command: &protocol.Command{
						Cmd:  "toggle",
						Args: []string{dev.Id},
					},
					Feedback: `Devices["` + dev.Id + `"].PrimaryApp`,
				})
			case "Updated":
				serverSendChannel <- node.Node()
			default:
				log.Warn("Unknown event: ", event.Name)
			}
		}
	}()

	chromecast.Listen()

	select {}
} /*}}}*/

// WORKER that monitors the current connection state
func monitorState(node *protocol.Node, connectionState chan int, send chan interface{}) {
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
