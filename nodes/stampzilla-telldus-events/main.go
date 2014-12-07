package main

import (
	"flag"
	"strconv"
	"unsafe"

	log "github.com/cihub/seelog"
	"github.com/stampzilla/stampzilla-go/nodes/basenode"
	"github.com/stampzilla/stampzilla-go/protocol"
)

/*
#cgo LDFLAGS: -ltelldus-core

#include <telldus-core.h>

extern void registerCallbacks();
extern void unregisterCallbacks();
extern int updateDevices();

*/
import "C"

var node *protocol.Node
var state *State = &State{[]*Device{}, make(map[string]*Sensor, 0)}

func main() {
	// Load logger
	//logger, err := log.LoggerFromConfigAsFile("../logconfig.xml")
	//if err != nil {
	//panic(err)
	//}
	//log.ReplaceLogger(logger)

	//Get a config with the correct parameters
	config := basenode.NewConfig()
	basenode.SetConfig(config)

	// Load flags
	//var host string
	//var port string
	//flag.StringVar(&host, "host", "localhost", "Stampzilla server hostname")
	//flag.StringVar(&port, "port", "8282", "Stampzilla server port")
	flag.Parse()

	log.Info("Starting TELLDUS-events node")

	C.registerCallbacks()
	defer C.unregisterCallbacks()

	// Create new node description
	node = protocol.NewNode("telldus-events")
	node.SetState(state)

	// Describe available actions
	node.AddAction("set", "Set", []string{"Devices.Id"})
	node.AddAction("toggle", "Toggle", []string{"Devices.Id"})
	node.AddAction("dim", "Dim", []string{"Devices.Id", "value"})

	// Describe available layouts
	//node.AddLayout("1", "switch", "toggle", "Devices", []string{"on"}, "Switches")
	//node.AddLayout("2", "slider", "dim", "Devices", []string{"dim"}, "Dimmers")
	//node.AddLayout("3", "slider", "dim", "Devices", []string{"dim"}, "Specials")

	// Add devices
	cnt := C.updateDevices()
	log.Info("Updated devices (", cnt, " in total)")

	for _, dev := range state.Devices {
		node.AddElement(&protocol.Element{
			Type: protocol.ElementTypeToggle,
			Name: dev.Name,
			Command: &protocol.Command{
				Cmd:  "toggle",
				Args: []string{dev.Id},
			},
			Feedback: `Devices[` + dev.Id + `].On`,
		})
	}

	// Start the connection
	//go connection(host, port, node)

	//Create channels so we can communicate with the stampzilla-go server
	serverSendChannel := make(chan interface{})
	serverRecvChannel := make(chan protocol.Command)
	connectionState := basenode.Connect(serverSendChannel, serverRecvChannel)
	go monitorState(connectionState, serverSendChannel)

	// This worker recives all incomming commands
	go serverRecv(serverRecvChannel)

	select {}
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

// WORKER that recives all incomming commands
func serverRecv(recv chan protocol.Command) {
	for d := range recv {
		processCommand(d)
	}
}

func processCommand(cmd protocol.Command) {
	var result C.int = C.TELLSTICK_ERROR_UNKNOWN
	var id C.int = 0

	i, err := strconv.Atoi(cmd.Args[0])
	if err != nil {
		log.Error("Failed to decode arg[0] to int", err, cmd.Args[0])
		return
	}

	id = C.int(i)

	switch cmd.Cmd {
	case "on":
		result = C.tdTurnOn(id)
	case "off":
		result = C.tdTurnOff(id)
	case "toggle":
		s := C.tdLastSentCommand(id, C.TELLSTICK_TURNON|C.TELLSTICK_TURNOFF|C.TELLSTICK_DIM)
		switch {
		case s&C.TELLSTICK_DIM != 0:
			var state *C.char = C.tdLastSentValue(id)
			log.Info("DIM: ", C.GoString(state))
			if C.GoString(state) == "0" {
				result = C.tdTurnOn(id)
			} else {
				result = C.tdTurnOff(id)
			}
			C.tdReleaseString(state)
		case s&C.TELLSTICK_TURNON != 0:
			result = C.tdTurnOff(id)
		case s&C.TELLSTICK_TURNOFF != 0:
			result = C.tdTurnOn(id)
		}
	}

	if result != C.TELLSTICK_SUCCESS {
		var errorString *C.char = C.tdGetErrorString(result)
		log.Error(C.GoString(errorString))
		C.tdReleaseString(errorString)
	}
}

//export newDevice
func newDevice(id int, name *C.char, methods, s int, value *C.char) {
	//log.Info(id, C.GoString(name))

	features := []string{}
	if methods&C.TELLSTICK_TURNON != 0 {
		features = append(features, "on")
	}
	if methods&C.TELLSTICK_TURNOFF != 0 {
		features = append(features, "off")
	}
	if methods&C.TELLSTICK_BELL != 0 {
		features = append(features, "bell")
	}
	if methods&C.TELLSTICK_TOGGLE != 0 {
		features = append(features, "toggle")
	}
	if methods&C.TELLSTICK_DIM != 0 {
		features = append(features, "dim")
	}
	if methods&C.TELLSTICK_EXECUTE != 0 {
		features = append(features, "execute")
	}
	if methods&C.TELLSTICK_UP != 0 {
		features = append(features, "up")
	}
	if methods&C.TELLSTICK_DOWN != 0 {
		features = append(features, "down")
	}
	if methods&C.TELLSTICK_STOP != 0 {
		features = append(features, "stop")
	}

	if s&C.TELLSTICK_TURNON != 0 {
		state.AddDevice(strconv.Itoa(id), C.GoString(name), features, DeviceState{On: true, Dim: 100})
	}
	if s&C.TELLSTICK_TURNOFF != 0 {
		state.AddDevice(strconv.Itoa(id), C.GoString(name), features, DeviceState{On: false})
	}
	if s&C.TELLSTICK_DIM != 0 {
		var currentState = C.GoString(value)
		level, _ := strconv.ParseUint(currentState, 10, 16)
		state.AddDevice(strconv.Itoa(id), C.GoString(name), features, DeviceState{On: level > 0, Dim: int(level)})
	}

}

//export sensorEvent
func sensorEvent(protocol, model *C.char, sensorId, dataType int, value *C.char) {
	log.Debugf("SensorEVENT %s,\t%s,\t%d -> ", C.GoString(protocol), C.GoString(model), sensorId)

	if dataType == C.TELLSTICK_TEMPERATURE {
		log.Debugf("Temperature:\t%s\n", C.GoString(value))
	} else if dataType == C.TELLSTICK_HUMIDITY {
		log.Debugf("Humidity:\t%s%%\n", C.GoString(value))
	}
}

//export deviceEvent
func deviceEvent(deviceId, method int, data *C.char, callbackId int, context unsafe.Pointer) {
	log.Debugf("DeviceEVENT %d\t%d\t:%s\n", deviceId, method, C.GoString(data))
}

//export deviceChangeEvent
func deviceChangeEvent(deviceId, changeEvent, changeType, callbackId int, context unsafe.Pointer) {
	log.Debugf("DeviceChangeEVENT %d\t%d\t%d\n", deviceId, changeEvent, changeType)
}

//export rawDeviceEvent
func rawDeviceEvent(data *C.char, controllerId, callbackId int, context unsafe.Pointer) {
	log.Debugf("rawDeviceEVENT (%d):%s\n", controllerId, C.GoString(data))
}
