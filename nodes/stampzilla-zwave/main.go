package main

import (
	"flag"

	log "github.com/cihub/seelog"
	"github.com/stampzilla/gozwave"
	"github.com/stampzilla/gozwave/commands/switchbinary"
	"github.com/stampzilla/stampzilla-go/nodes/basenode"
	"github.com/stampzilla/stampzilla-go/pkg/notifier"
	"github.com/stampzilla/stampzilla-go/protocol"
)

var VERSION string = "dev"
var BUILD_DATE string = ""

// MAIN - This is run when the init function is done

var notify *notifier.Notify

func main() { /*{{{*/
	log.Info("Starting ZWAVE node")

	// Parse all commandline arguments, host and port parameters are added in the basenode init function
	flag.Parse()

	//Get a config with the correct parameters
	config := basenode.NewConfig()

	//Activate the config
	basenode.SetConfig(config)

	z, err := gozwave.Connect("/dev/ttyACM0", "zwave-networkmap.json")
	if err != nil {
		log.Error(err)
		return
	}

	node := protocol.NewNode("zwave")
	node.Version = VERSION
	node.BuildDate = BUILD_DATE

	//Start communication with the server
	connection := basenode.Connect()
	notify = notifier.New(connection)
	notify.SetSource(node)

	// Thit worker keeps track on our connection state, if we are connected or not
	go monitorState(node, connection)

	node.AddElement(&protocol.Element{
		Type: protocol.ElementTypeButton,
		Name: "Up",
		Command: &protocol.Command{
			Cmd:  "blinds",
			Args: []string{"0"},
		},
	})
	node.AddElement(&protocol.Element{
		Type: protocol.ElementTypeButton,
		Name: "Down",
		Command: &protocol.Command{
			Cmd:  "blinds",
			Args: []string{"1"},
		},
	})

	state := NewState()
	node.SetState(state)
	state.zwave = z

	// This worker recives all incomming commands
	go serverRecv(node, connection)

	//state.Nodes, _ = z.GetNodes()
	//connection.Send(node.Node())

	z.GetNodes()

	for {
		select {
		case event := <-z.GetNextEvent():
			log.Infof("Event: %#v", event)
		}
	}
} /*}}}*/

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
func serverRecv(node *protocol.Node, connection basenode.Connection) {
	for d := range connection.Receive() {
		processCommand(node, connection, d)
	}
}

// THis is called on each incomming command
func processCommand(node *protocol.Node, connection basenode.Connection, cmd protocol.Command) {
	if s, ok := node.State().(*State); ok {
		log.Infof("Incoming command from server: %#v \n", cmd, s)
		if len(cmd.Args) == 0 {
			return
		}

		switch cmd.Cmd {
		case "blinds":

			rollup := switchbinary.New().SetNode(2)

			if cmd.Args[0] == "1" {
				rollup.SetValue(true)
			}

			<-s.zwave.Send(rollup) // Stop previous motion
			//<-time.After(time.Millisecond * 200)
			//<-s.zwave.Send(rollup) // Start up
			//connection.Send(node.Node())
		}
	}
}
