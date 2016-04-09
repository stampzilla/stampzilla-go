package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"strconv"
	"unsafe"

	"github.com/stampzilla/stampzilla-go/nodes/basenode"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-telldus-events/sensormonitor"
	"github.com/stampzilla/stampzilla-go/pkg/notifier"
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
var state *State = &State{make(map[string]*Device), make(map[string]*Sensor, 0)}
var serverConnection basenode.Connection
var sensorMonitor sensormonitor.Monitor

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

	log.Println("Starting TELLDUS-events node")

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
	log.Println("Updated devices (", cnt, " in total)")

	for _, dev := range state.Devices {
		node.AddElement(&protocol.Element{
			Type: protocol.ElementTypeToggle,
			Name: dev.Name,
			Command: &protocol.Command{
				Cmd:  "toggle",
				Args: []string{dev.Id},
			},
			Feedback: `Devices[` + dev.Id + `].State.On`,
		})
	}

	// Start the connection
	//go connection(host, port, node)

	serverConnection = basenode.Connect()
	notify := notifier.New(serverConnection)
	notify.SetSource(node)

	sensorMonitor = sensormonitor.New(notify)
	sensorMonitor.Start()

	go monitorState(serverConnection)

	// This worker recives all incomming commands
	go serverRecv(serverConnection)

	select {}
}

// WORKER that monitors the current connection state
func monitorState(connection basenode.Connection) {
	for s := range connection.State() {
		switch s {
		case basenode.ConnectionStateConnected:
			connection.Send(node.Node())
		case basenode.ConnectionStateDisconnected:
		}
	}
}

// WORKER that recives all incomming commands
func serverRecv(connection basenode.Connection) {
	send := processCommandWorker()
	for d := range connection.Receive() {
		send <- d
	}
}

func processCommandWorker() chan protocol.Command {
	var send = make(chan protocol.Command, 100)

	go func() {
		for c := range send {
			if err := processCommand(c); err != nil {
				log.Println(err)
			}
		}
	}()

	return send
}

func processCommand(cmd protocol.Command) error {
	log.Println("Processing command", cmd)
	var result C.int = C.TELLSTICK_ERROR_UNKNOWN
	var id C.int = 0

	i, err := strconv.Atoi(cmd.Args[0])
	if err != nil {
		return fmt.Errorf("Failed to decode arg[0] to int %s %s", err, cmd.Args[0])
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
			log.Println("DIM: ", C.GoString(state))
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
		C.tdReleaseString(errorString)
		return errors.New(C.GoString(errorString))
	}

	return nil
}

//export newDevice
func newDevice(id int, name *C.char, methods, s int, value *C.char) {
	//log.Println(id, C.GoString(name))

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
	//log.Println("SensorEVENT: ", C.GoString(protocol), C.GoString(model), sensorId)

	var s *Sensor
	if s = state.GetSensor(sensorId); s == nil {
		s = state.AddSensor(sensorId, "UNKNOWN")
	}
	sensorMonitor.Alive(s.Id)

	if dataType == C.TELLSTICK_TEMPERATURE {
		t, _ := strconv.ParseFloat(C.GoString(value), 64)
		log.Printf("Temperature %d : %f\n", s.Id, t)
		if s.Temp != t {
			//log.Println("Difference, sending to server")
			s.Temp = t
			serverConnection.Send(node.Node())
		}
	} else if dataType == C.TELLSTICK_HUMIDITY {
		h, _ := strconv.ParseFloat(C.GoString(value), 64)
		log.Printf("Humidity %d : %f\n", s.Id, h)
		if s.Humidity != h {
			//log.Println("Difference, sending to server")
			s.Humidity = h
			serverConnection.Send(node.Node())
		}
	}
}

//export deviceEvent
func deviceEvent(deviceId, method int, data *C.char, callbackId int, context unsafe.Pointer) {
	//log.Println("DeviceEVENT: ", deviceId, method, C.GoString(data))
	device := state.GetDevice(strconv.Itoa(deviceId))
	if method&C.TELLSTICK_TURNON != 0 {
		device.State.On = true
		serverConnection.Send(node.Node())
	}
	if method&C.TELLSTICK_TURNOFF != 0 {
		device.State.On = false
		serverConnection.Send(node.Node())
	}
	if method&C.TELLSTICK_DIM != 0 {
		level, err := strconv.ParseUint(C.GoString(data), 10, 16)
		if err != nil {
			log.Println(err)
			return
		}
		if level == 0 {
			device.State.On = false
		}
		if level > 0 {
			device.State.On = true
		}
		device.State.Dim = int(level)
		serverConnection.Send(node.Node())
	}
}

//export deviceChangeEvent
func deviceChangeEvent(deviceId, changeEvent, changeType, callbackId int, context unsafe.Pointer) {
	//log.Println("DeviceChangeEVENT: ", deviceId, changeEvent, changeType)
}

//export rawDeviceEvent
func rawDeviceEvent(data *C.char, controllerId, callbackId int, context unsafe.Pointer) {
	//log.Println("rawDeviceEVENT: ", controllerId, C.GoString(data))
}
