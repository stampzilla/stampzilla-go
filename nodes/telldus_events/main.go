package main

import (
	"flag"
	"fmt"
	log "github.com/cihub/seelog"
	"github.com/stampzilla/stampzilla-go/protocol"
	"strconv"
	"unsafe"
)

/*
#cgo LDFLAGS: -ltelldus-core

#include <telldus-core.h>

extern void registerCallbacks();
extern void unregisterCallbacks();
extern void updateDevices();

*/
import "C"

var node *protocol.Node

func main() {
	// Load logger
	logger, err := log.LoggerFromConfigAsFile("../logconfig.xml")
	if err != nil {
		panic(err)
	}
	log.ReplaceLogger(logger)

	// Load flags
	var host string
	var port string
	flag.StringVar(&host, "host", "localhost", "Stampzilla server hostname")
	flag.StringVar(&port, "port", "8282", "Stampzilla server port")
	flag.Parse()

	log.Info("Starting TELLDUS-events node")

	C.registerCallbacks()
	defer C.unregisterCallbacks()

	// Create new node description
	node = protocol.NewNode("tellstick-event")

	// Describe available actions
	node.AddAction("set", "Set", []string{"Devices.Id"})
	node.AddAction("toggle", "Toggle", []string{"Devices.Id"})
	node.AddAction("dim", "Dim", []string{"Devices.Id", "value"})

	// Describe available layouts
	node.AddLayout("1", "switch", "toggle", "Devices", []string{"on"}, "Switches")
	node.AddLayout("2", "slider", "dim", "Devices", []string{"dim"}, "Dimmers")
	node.AddLayout("3", "slider", "dim", "Devices", []string{"dim"}, "Specials")

	// Add devices
	C.updateDevices()

	// Start the connection
	go connection(host, port, node)

	select {}
}

func processCommand(cmd protocol.Command) {
	var result C.int = C.TELLSTICK_ERROR_UNKNOWN
	var id C.int = 0

	i, err := strconv.Atoi(cmd.Args[0])
	if err != nil {
		log.Error(err)
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

	var state string

	if s&C.TELLSTICK_TURNON != 0 {
		state = "true"
	}
	if s&C.TELLSTICK_TURNOFF != 0 {
		state = "false"
	}
	if s&C.TELLSTICK_DIM != 0 {
		state = C.GoString(value)
	}

	node.AddDevice(strconv.Itoa(id), C.GoString(name), features, state)
}

//export sensorEvent
func sensorEvent(protocol, model *C.char, sensorId, dataType int, value *C.char) {
	fmt.Printf("SensorEVENT %s,\t%s,\t%d -> ", C.GoString(protocol), C.GoString(model), sensorId)

	if dataType == C.TELLSTICK_TEMPERATURE {
		fmt.Printf("Temperature:\t%s\n", C.GoString(value))
	} else if dataType == C.TELLSTICK_HUMIDITY {
		fmt.Printf("Humidity:\t%s%%\n", C.GoString(value))
	}
}

//export deviceEvent
func deviceEvent(deviceId, method int, data *C.char, callbackId int, context unsafe.Pointer) {
	fmt.Printf("DeviceEVENT %d\t%d\t:%s\n", deviceId, method, C.GoString(data))
}

//export deviceChangeEvent
func deviceChangeEvent(deviceId, changeEvent, changeType, callbackId int, context unsafe.Pointer) {
	fmt.Printf("DeviceChangeEVENT %d\t%d\t%d\n", deviceId, changeEvent, changeType)
}

//export rawDeviceEvent
func rawDeviceEvent(data *C.char, controllerId, callbackId int, context unsafe.Pointer) {
	fmt.Printf("rawDeviceEVENT (%d):%s\n", controllerId, C.GoString(data))
}
