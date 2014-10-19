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

var node *protocol.Node
var state *State
var serverSendChannel chan interface{}
var serverRecvChannel chan protocol.Command

func init() {
	flag.Parse()

	//Setup Config
	config := basenode.NewConfig()
	basenode.SetConfig(config)

	//Start communication with the server
	serverSendChannel = make(chan interface{})
	serverRecvChannel = make(chan protocol.Command)
	connectionState := basenode.Connect(serverSendChannel, serverRecvChannel)
	go monitorState(connectionState, serverSendChannel)
	go serverRecv(serverRecvChannel)

}

func main() {

	node = protocol.NewNode("enocean")
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

	setupEnoceanCommunication()
}

func monitorState(connectionState chan int, send chan interface{}) {
	for s := range connectionState {
		switch s {
		case basenode.ConnectionStateConnected:
			send <- node
		case basenode.ConnectionStateDisconnected:
		}
	}
}

func serverRecv(recv chan protocol.Command) {

	for d := range recv {
		processCommand(d)
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

func setupEnoceanCommunication() {
	enoceanSend = make(chan goenocean.Encoder)
	recv := make(chan goenocean.Packet)
	goenocean.Serial(enoceanSend, recv)

	reciever(recv)
}

func reciever(recv chan goenocean.Packet) {
	for p := range recv {
		if p.SenderId() != [4]byte{0, 0, 0, 0} {
			incomingPacket(p)
		}
	}
}

func incomingPacket(p goenocean.Packet) {

	var d *Device
	if d = state.Device(p.SenderId()); d == nil {
		//Add unknown device
		d = state.AddDevice(p.SenderId(), "UNKNOWN", nil, false)
		saveDevicesToFile()
		serverSendChannel <- node
	}

	if t, ok := p.(goenocean.Telegram); ok {
		for _, deviceEep := range d.RecvEEPs {
			if deviceEep[0:2] != hex.EncodeToString([]byte{t.TelegramType()}) {
				continue
			}

			if h := handlers.getHandler(deviceEep); h != nil {
				h.Process(d, t)
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
