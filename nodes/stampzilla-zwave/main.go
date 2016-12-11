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
	"github.com/stampzilla/gozwave/nodes"
	"github.com/stampzilla/stampzilla-go/nodes/basenode"
	"github.com/stampzilla/stampzilla-go/pkg/notifier"
	"github.com/stampzilla/stampzilla-go/protocol"
	"github.com/stampzilla/stampzilla-go/protocol/devices"
)

var VERSION string = "dev"
var BUILD_DATE string = ""

// MAIN - This is run when the init function is done

var notify *notifier.Notify

func main() {
	log.Info("Starting ZWAVE node")

	debug := flag.Bool("v", false, "Verbose - show more debuging info")
	port := flag.String("controllerport", "/dev/ttyACM0", "SerialAPI communication port (to controller)")

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

	z, err := gozwave.Connect(*port, "zwave-networkmap.json")
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

	state := NewState()
	node.SetState(state)
	state.zwave = z

	// This worker recives all incomming commands
	go serverRecv(node, connection)

	<-time.After(time.Second) // TODO: Wait for node.Uuid_ to be populated

	// Add all existing nodes to the state / device list
	for _, znode := range z.Nodes.All() {
		if znode.Id == 1 {
			continue
		}

		state.Nodes = append(state.Nodes, newZwavenode(znode))
		n := state.GetNode(znode.Id)
		n.sync(znode)

		addOrUpdateDevice(node, znode)
	}
	connection.Send(node.Node())

	// Listen from events from the zwave-controller
	for {
		select {
		case event := <-z.GetNextEvent():
			log.Infof("Event: %#v", event)
			switch e := event.(type) {
			case events.NodeDiscoverd:
				znode := z.Nodes.Get(e.Address)
				log.Infof("%#v", znode)
				if znode != nil {
					state.Nodes = append(state.Nodes, newZwavenode(znode))
				}

			case events.NodeUpdated:
				n := state.GetNode(e.Address)
				if n != nil {
					znode := z.Nodes.Get(e.Address)

					addOrUpdateDevice(node, znode)
					n.sync(znode)
				}
			}

			connection.Send(node.Node())
		}
	}
}

func addOrUpdateDevice(node *protocol.Node, znode *nodes.Node) {
	if znode.Device == nil {
		return
	}

	log.Errorf("Endpoints: %#v", znode.Endpoints)

	for i := 0; i < len(znode.Endpoints); i++ {
		devid := strconv.Itoa(int(znode.Id) + (i * 1000))

		switch {
		case znode.HasCommand(commands.SwitchMultilevel):
			node.Devices().Add(&devices.Device{
				Type:     "dimmableLamp",
				Name:     znode.Device.Brand + " - " + znode.Device.Product + " (Address: " + devid + ")",
				Id:       devid,
				Online:   true,
				Node:     node.Uuid(),
				StateMap: map[string]string{
				//TODO add state map
				//"On": "Devices[" + devid + "]" + ".State.On",
				},
			})
		case znode.HasCommand(commands.SwitchBinary):
			node.Devices().Add(&devices.Device{
				Type:     "lamp",
				Name:     znode.Device.Brand + " - " + znode.Device.Product + " (Address: " + devid + ")",
				Id:       devid,
				Online:   true,
				Node:     node.Uuid(),
				StateMap: map[string]string{
				//TODO add state map
				//"On": "Devices[" + devid + "]" + ".State.On",
				},
			})
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

		var device gozwave.Controllable

		endpoint := int(id / 1000)
		id = id - (endpoint * 1000)

		znode := s.zwave.Nodes.Get(id)
		if znode == nil {
			log.Error("Node not found")
			return
		}

		if id < 1000 && len(znode.Endpoints) < 2 {
			device = znode
		} else {
			device = znode.Endpoint(endpoint)
		}

		switch cmd.Cmd {
		case "on":
			device.On()
		case "off":
			device.Off()
		case "level":
			level, err := strconv.ParseFloat(cmd.Args[1], 64)
			if err != nil {
				log.Error(err)
				return
			}
			device.Level(level)
		default:
			log.Warnf("Unknown command '%s'", cmd.Cmd)
		}
	}
}
