package main

import (
	//"github.com/bjeanes/go-lifx"

	"flag"
	"strconv"
	"strings"
	"time"

	log "github.com/cihub/seelog"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/stamp/go-lifx/client"
	"github.com/stampzilla/stampzilla-go/nodes/basenode"
	"github.com/stampzilla/stampzilla-go/protocol"
)

var VERSION string = "dev"
var BUILD_DATE string = ""

// GLOBAL VARS
var node = &Node{}
var state *State
var serverSendChannel chan interface{}
var serverRecvChannel chan protocol.Command
var lifxClient *client.Client

// Config contains the lifx specific node configuration
type Config struct {
	BroadcastAddress string `json:"broadcastAddress"`
	APIAccessToken   string `json:"api_access_token"`
}

// Node is an extended version of protocol.Node with a pointer to the connection
type Node struct {
	*protocol.Node
	connection basenode.Connection
}

func main() {
	// Parse all commandline arguments, host and port parameters are added in the basenode init function
	flag.Parse()

	//Get a config with the correct parameters
	config := basenode.NewConfig()

	//Activate the config
	basenode.SetConfig(config)
	nc := &Config{}
	err := config.NodeSpecific(&nc)
	if err != nil {
		log.Error(err)
		//return
	}

	node.Node = protocol.NewNode("lifx")
	node.Version = VERSION
	node.BuildDate = BUILD_DATE

	//Start communication with the server
	node.connection = basenode.Connect()

	// This worker keeps track on our connection state, if we are connected or not
	go monitorState(node, node.connection)

	// This worker recives all incomming commands
	go serverRecv(node.connection)

	// Create new node description
	state = NewState()
	state.publishStateFunc = node.publishState
	node.SetState(state)

	if nc.BroadcastAddress != "" {
		lifxClient, err = client.NewWithBroadcastAddress(nc.BroadcastAddress)
	} else {
		lifxClient, err = client.New()
	}
	if err != nil {
		log.Error(err)
		return
	}

	if nc.APIAccessToken != "" {
		lifxCloudClient, err := newLifxCloudClient(nc.APIAccessToken)
		lifxCloudClient.pollerResultFunc = node.handleLifxCloudUpdate
		if err != nil {
			log.Error(err)
		} else {
			lifxCloudClient.start()
		}
	}

	go discoverWorker(lifxClient, node.connection)

	select {}

}

func (node *Node) publishState() {
	node.connection.Send(node.Node)
}

func discoverWorker(client *client.Client, connection basenode.Connection) {
	go monitorLampCollection(client.Lights, connection)

	for {
		log.Info("Try to discover bulbs:")
		for _ = range client.Discover() {
		}

		<-time.After(60 * time.Second)
	}
}

// WORKER that monitors the list of lamps
func monitorLampCollection(lights *client.LightCollection, connection basenode.Connection) {
	for s := range lights.Lights.Changes {
		switch s.Event {
		case client.LampAdded:
			log.Warnf("Added: %s (%s)", s.Lamp.Id(), s.Lamp.Label())

			state.AddLanDevice(s.Lamp)

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
			//log.Infof("Collection: %#v", lights.Lights)
		case client.LampUpdated:
			log.Warnf("Changed: %s (%s)", s.Lamp.Id(), s.Lamp.Label())
		case client.LampRemoved:
			log.Warnf("Removed: %s (%s)", s.Lamp.Id(), s.Lamp.Label())
		default:
			log.Infof("Received unknown event: %d", s.Event)
		}
	}
}

func (node *Node) handleLifxCloudUpdate(lamps []*cloudGetAllResponse) {
	for _, v := range lamps {
		if len(v.ID) == 12 {
			v.ID = v.ID + "0000"
		}

		d := state.GetByID(v.ID)
		if d == nil {
			d = NewLamp(v.ID, v.Label, "")
			state.Add(d)
		}

		d.SyncFromCloud(v)
	}
}

// WORKER that monitors the current connection state
func monitorState(node *Node, connection basenode.Connection) {
	for s := range connection.State() {
		switch s {
		case basenode.ConnectionStateConnected:
			node.publishState()
		case basenode.ConnectionStateDisconnected:
		}
	}
}

// WORKER that recives all incoming commands
func serverRecv(connection basenode.Connection) {
	for d := range connection.Receive() {
		processCommand(d)
	}
}

// This is called on each incoming command
func processCommand(cmd protocol.Command) {
	log.Info("Incoming command from server:", cmd)

	if len(cmd.Args) < 1 {
		return
	}

	lamp, err := lifxClient.Lights.GetById(cmd.Args[0])
	if err != nil {
		log.Error(err)
		return
	}

	switch cmd.Cmd {
	case "set":
		lamp.TurnOn()
		if len(cmd.Args) > 1 {
			if strings.Index(cmd.Args[1], "#") != 0 {
				cmd.Args[1] = "#" + cmd.Args[1]
			}

			c, err := colorful.Hex(cmd.Args[1])
			if err != nil {
				log.Error("Failed to decode color: ", err)
				return
			}

			h, s, v := c.Hsv()

			if len(cmd.Args) > 2 {
				duration, err := strconv.Atoi(cmd.Args[2])
				if err == nil {
					if len(cmd.Args) > 3 {
						kelvin, err := strconv.Atoi(cmd.Args[3])
						if err == nil {
							lamp.SetState(h, s, v, uint32(duration), uint32(kelvin))
							return
						}
					}
					lamp.SetState(h, s, v, uint32(duration), 6500)
					return
				}
				log.Error("Failed to decode duration: ", err)
			}

			lamp.SetState(h, s, v, 1000, 6500)
		}
	case "on":
		lamp.TurnOn()
	case "off":
		lamp.TurnOff()
	case "color":
		c, err := colorful.Hex(cmd.Params[0])
		if err != nil {
			log.Error(err)
			return
		}

		h, s, v := c.Hsv()

		lamp.SetState(h, s, v, 1000, 6500)
	}
}
