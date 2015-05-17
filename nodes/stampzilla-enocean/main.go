// Package main provides ...
package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"flag"
	"os"
	"strconv"
	"sync"

	log "github.com/cihub/seelog"
	"github.com/jonaz/goenocean"
	"github.com/stampzilla/stampzilla-go/nodes/basenode"
	"github.com/stampzilla/stampzilla-go/protocol"
)

var state *State

func main() {

	node := protocol.NewNode("enocean")

	flag.Parse()

	//Setup Config
	config := basenode.NewConfig()
	basenode.SetConfig(config)

	//Start communication with the server
	//serverSendChannel = make(chan interface{})
	//serverRecvChannel = make(chan protocol.Command)
	connection := basenode.Connect()
	go monitorState(node, connection)
	go serverRecv(connection)

	// Describe available actions
	node.AddAction("set", "Set", []string{"Devices.Id"})
	node.AddAction("toggle", "Toggle", []string{"Devices.Id"})
	node.AddAction("dim", "Dim", []string{"Devices.Id", "value"})

	// Describe available layouts
	node.AddLayout("1", "switch", "toggle", "Devices", []string{"on"}, "Switches")
	node.AddLayout("2", "slider", "dim", "Devices", []string{"dim"}, "Dimmers")
	node.AddLayout("3", "slider", "dim", "Devices", []string{"dim"}, "Specials")

	node.AddElement(&protocol.Element{
		Type: protocol.ElementTypeToggle,
		Name: "Lamp 0186ff7d",
		Command: &protocol.Command{
			Cmd:  "toggle",
			Args: []string{"0186ff7d"},
		},
		Feedback: `Devices["0186ff7d"].On`,
	})
	node.AddElement(&protocol.Element{
		Type: protocol.ElementTypeText,
		Name: "Lamp 0186ff7d power",
		//Command: &protocol.Command{
		//Cmd:  "toggle",
		//Args: []string{"0186ff7d"},
		//},
		Feedback: `Devices["0186ff7d"].PowerW`,
	})

	//Setup state
	state = NewState()
	state.Devices = readConfigFromFile()
	node.SetState(state)

	elementGenerator := &ElementGenerator{}
	elementGenerator.State = state
	elementGenerator.Node = node
	elementGenerator.Run()

	checkDuplicateSenderIds()

	setupEnoceanCommunication(node, connection)
}

func monitorState(node *protocol.Node, connection *basenode.Connection) {
	for s := range connection.State {
		switch s {
		case basenode.ConnectionStateConnected:
			connection.Send <- node
		case basenode.ConnectionStateDisconnected:
		}
	}
}

func serverRecv(connection *basenode.Connection) {

	for d := range connection.Receive {
		processCommand(d)
	}

}

func checkDuplicateSenderIds() {
	for _, d := range state.Devices {
		id1 := d.Id()[3] & 0x7f
		for _, d1 := range state.Devices {
			if d.Id() == d1.Id() {
				continue
			}
			id2 := d1.Id()[3] & 0x7f
			if id2 == id1 {
				log.Error("DUPLICATE ID FOUND when generating senderIds for eltako devices")
			}
		}
	}
}

func processCommand(cmd protocol.Command) {
	log.Debug("INCOMING COMMAND", cmd)
	if len(cmd.Args) == 0 {
		log.Error("Missing device ID in arguments")
		return
	}

	device := state.DeviceByString(cmd.Args[0])
	switch cmd.Cmd {
	case "toggle":
		device.CmdToggle()
	case "on":
		device.CmdOn()
	case "off":
		device.CmdOff()
	case "dim":
		lvl, _ := strconv.Atoi(cmd.Args[1])
		device.CmdDim(lvl)
	case "learn":
		device.CmdLearn()
	}
}

var enoceanSend chan goenocean.Encoder

func setupEnoceanCommunication(node *protocol.Node, connection *basenode.Connection) {

	enoceanSend = make(chan goenocean.Encoder)
	recv := make(chan goenocean.Packet)
	goenocean.Serial(enoceanSend, recv)

	getIDBase()
	reciever(node, connection, recv)
}

func getIDBase() {
	p := goenocean.NewPacket()
	p.SetPacketType(goenocean.PacketTypeCommonCommand)
	p.SetData([]byte{0x08})
	enoceanSend <- p
}

var usb300SenderId [4]byte

func reciever(node *protocol.Node, connection *basenode.Connection, recv chan goenocean.Packet) {
	for p := range recv {
		if p.PacketType() == goenocean.PacketTypeResponse && len(p.Data()) == 5 {
			copy(usb300SenderId[:], p.Data()[1:4])
			log.Debugf("senderid: % x", usb300SenderId)
			continue
		}
		if p.SenderId() != [4]byte{0, 0, 0, 0} {
			incomingPacket(node, connection, p)
		}
	}
}

func incomingPacket(node *protocol.Node, connection *basenode.Connection, p goenocean.Packet) {

	var d *Device
	if d = state.Device(p.SenderId()); d == nil {
		//Add unknown device
		d = state.AddDevice(p.SenderId(), "UNKNOWN", nil, false)
		saveDevicesToFile()
		connection.Send <- node
	}

	log.Debug("Incoming packet")
	if t, ok := p.(goenocean.Telegram); ok {
		log.Debug("Packet is goenocean.Telegram")
		for _, deviceEep := range d.RecvEEPs {
			if deviceEep[0:2] != hex.EncodeToString([]byte{t.TelegramType()}) {
				log.Debug("Packet is wrong deviceEep ", deviceEep, t.TelegramType())
				continue
			}

			if h := handlers.getHandler(deviceEep); h != nil {
				h.Process(d, t)
				log.Info("Incoming packet processed from", d.IdString())
				//TODO add return bool in process and to send depending on that!
				connection.Send <- node
				return
			}
		}
	}

	//fmt.Println("Unknown packet")

}

var devFileMutex sync.Mutex

func saveDevicesToFile() {
	devFileMutex.Lock()
	defer devFileMutex.Unlock()
	configFile, err := os.Create("devices.json")
	if err != nil {
		log.Error("creating config file", err.Error())
	}
	var out bytes.Buffer
	b, err := json.MarshalIndent(state.Devices, "", "\t")
	if err != nil {
		log.Error("error marshal json", err)
	}
	json.Indent(&out, b, "", "\t")
	out.WriteTo(configFile)
}
func readConfigFromFile() map[string]*Device {
	devFileMutex.Lock()
	defer devFileMutex.Unlock()
	configFile, err := os.Open("devices.json")
	if err != nil {
		log.Error("opening config file", err.Error())
	}

	config := make(map[string]*Device)
	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&config); err != nil {
		log.Error("parsing config file", err.Error())
	}

	return config
}
