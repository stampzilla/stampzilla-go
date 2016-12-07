package main

import (
	"flag"
	"strconv"
	"time"

	"github.com/Sirupsen/logrus"
	log "github.com/cihub/seelog"
	"github.com/stampzilla/gozwave"
	"github.com/stampzilla/gozwave/commands"
	"github.com/stampzilla/gozwave/events"
	"github.com/stampzilla/stampzilla-go/nodes/basenode"
	"github.com/stampzilla/stampzilla-go/pkg/notifier"
	"github.com/stampzilla/stampzilla-go/protocol"
)

var VERSION string = "dev"
var BUILD_DATE string = ""

// MAIN - This is run when the init function is done

var notify *notifier.Notify

func main() {
	log.Info("Starting ZWAVE node")

	debug := flag.Bool("v", false, "Verbose - show more debuging info")

	// Parse all commandline arguments, host and port parameters are added in the basenode init function
	flag.Parse()
	logrus.SetLevel(logrus.WarnLevel)
	if *debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

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

	for {
		select {
		case event := <-z.GetNextEvent():
			log.Infof("Event: %#v", event)
			switch e := event.(type) {
			case events.NodeDiscoverd:
				log.Infof("%#v", z.Nodes.Get(e.Address))
				state.Nodes = append(state.Nodes, z.Nodes.Get(e.Address))
			}

			connection.Send(node.Node())
		}
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

		id, err := strconv.Atoi(cmd.Args[0])
		if err != nil {
			log.Error(err)
			return
		}

		//device := s.zwave.Nodes.Get(byte(id))

		switch cmd.Cmd {
		case "on":
			//TODO SwitchBinary not working yet :(
			cmd := commands.NewSwitchBinary()
			cmd.SetValue(true)
			cmd.SetNode(byte(id))
			s.zwave.Connection.Send(cmd, time.Second)
		case "off":
			cmd := commands.NewSwitchBinary()
			cmd.SetValue(false)
			cmd.SetNode(byte(id))
			s.zwave.Connection.Send(cmd, time.Second)
		case "blinds":

			//rollup := switchbinary.New().SetNode(2)

			//if cmd.Args[0] == "1" {
			//	rollup.SetValue(true)
			//}

			//<-s.zwave.Send(rollup) // Stop previous motion
			//<-time.After(time.Millisecond * 200)
			//<-s.zwave.Send(rollup) // Start up
			//connection.Send(node.Node())
		}
	}
}
