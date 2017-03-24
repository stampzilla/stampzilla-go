package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	log "github.com/cihub/seelog"
	"github.com/stampzilla/stampzilla-go/nodes/basenode"
	"github.com/stampzilla/stampzilla-go/protocol"

	"github.com/stampzilla/gocast/discovery"
)

var VERSION string = "dev"
var BUILD_DATE string = ""

var state State

func main() {
	printVersion := flag.Bool("version", false, "Prints current version")
	if *printVersion != false {
		fmt.Println(VERSION + " (" + BUILD_DATE + ")")
		os.Exit(0)
	}

	// Parse all commandline arguments, host and port parameters are added in the basenode init function
	flag.Parse()

	log.Info("Starting CHROMECAST node")

	//Get a config with the correct parameters
	config := basenode.NewConfig()

	//Activate the config
	basenode.SetConfig(config)

	node := protocol.NewNode("chromecast")
	node.Version = VERSION
	node.BuildDate = BUILD_DATE

	//Start communication with the server
	connection := basenode.Connect()

	// Thit worker keeps track on our connection state, if we are connected or not
	go monitorState(node, connection)

	// This worker recives all incomming commands
	go serverRecv(connection.Receive())

	state = State{
		connection: connection,
		node:       node,
	}
	node.SetState(&state.Devices)

	discovery := discovery.NewService()

	go discoveryListner(discovery)
	discovery.Periodic(time.Second * 10)

	select {}
}

func discoveryListner(discovery *discovery.Service) {
	for device := range discovery.Found() {
		log.Debugf("New device discoverd: %s", device.String())
		NewChromecast(device)
		go func() {
			err := device.Connect()
			if err != nil {
				log.Error(err)
			}
		}()
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
	if len(cmd.Args) == 0 {
		log.Error("Missing argument 0 (which player?)")
		return
	}
	player := state.GetByUUID(cmd.Args[0])
	if player == nil {
		log.Errorf("Player with id %s not found", cmd.Args[0])
		return
	}

	switch cmd.Cmd {
	case "play":
		if len(cmd.Args) > 1 && strings.HasPrefix(cmd.Args[1], "http") {
			player.PlayUrl(strings.Join(cmd.Args[1:], "/"), "")
			return
		}
		player.Play()
	case "pause":
		player.Pause()
	case "stop":
		player.Stop()
	}
}
