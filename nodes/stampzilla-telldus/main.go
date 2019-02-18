package main

import (
	"errors"
	"strconv"
	"unsafe"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models/devices"
	n "github.com/stampzilla/stampzilla-go/pkg/node"
)

/*
#cgo LDFLAGS: -ltelldus-core

#include <telldus-core.h>

extern void registerCallbacks();
extern void unregisterCallbacks();
extern int updateDevices();

*/
import "C"

var node *n.Node

//var state *State = &State{make(map[string]*Device), make(map[string]*Sensor, 0)}

func main() {
	node = n.New("telldus")

	err := node.Connect()
	if err != nil {
		logrus.Error(err)
		return
	}

	//node.OnConfig(updatedConfig)

	C.registerCallbacks()
	defer C.unregisterCallbacks()

	cnt := C.updateDevices()
	logrus.Info("Updated ", cnt, " devices in total)")

	node.OnRequestStateChange(func(state devices.State, device *devices.Device) error {
		var result C.int = C.TELLSTICK_ERROR_UNKNOWN
		i, err := strconv.Atoi(device.ID.ID)
		if err != nil {
			return err
		}

		id := C.int(i)
		state.Bool("on", func(on bool) {
			if on {
				result = C.tdTurnOn(id)
				logrus.Debugf("turning on id %s\n", device.ID.ID)
				return
			}
			result = C.tdTurnOff(id)
			logrus.Debugf("turning off id %s\n", device.ID.ID)
		})
		if result != C.TELLSTICK_SUCCESS {
			var errorString *C.char = C.tdGetErrorString(result)
			C.tdReleaseString(errorString)
			err = errors.New(C.GoString(errorString))
		}
		return err
	})

	//notify := notifier.New(serverConnection)
	//notify.SetSource(node)

	//sensorMonitor = sensormonitor.New(notify)
	//sensorMonitor.MonitorSensors = nc.MonitorSensors
	//sensorMonitor.Start()
	//log.Println("Monitoring Sensors: ", nc.MonitorSensors)

	select {}
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

	strID := strconv.Itoa(id)
	dev := &devices.Device{
		Type:   "light",
		Name:   strID,
		ID:     devices.ID{ID: strID},
		Online: true,
		Traits: []string{"OnOff"},
		State: devices.State{
			"on":         false,
			"brightness": 0.0,
		},
	}

	if s&C.TELLSTICK_TURNON != 0 {
		//state.AddDevice(strconv.Itoa(id), C.GoString(name), features, DeviceState{On: true, Dim: 100})
		dev.State["on"] = true
		dev.State["brightness"] = 1.0
	}
	if s&C.TELLSTICK_TURNOFF != 0 {
		//state.AddDevice(strconv.Itoa(id), C.GoString(name), features, DeviceState{On: false})
		dev.State["on"] = false
		dev.State["brightness"] = 0.0
	}
	if s&C.TELLSTICK_DIM != 0 {
		var currentState = C.GoString(value)
		level, _ := strconv.ParseUint(currentState, 10, 16)
		//state.AddDevice(strconv.Itoa(id), C.GoString(name), features, DeviceState{On: level > 0, Dim: int(level)})
		dev.State["brightness"] = float64(level) / 255.0
		dev.State["on"] = level > 0
	}
	node.AddOrUpdate(dev)
}

//export sensorEvent
func sensorEvent(protocol, model *C.char, sensorID, dataType int, value *C.char) {
	//log.Println("SensorEVENT: ", C.GoString(protocol), C.GoString(model), sensorID)

	id := strconv.Itoa(sensorID)
	dev := node.GetDevice(id)
	if dev == nil {
		dev = &devices.Device{
			Type:   "sensor",
			Name:   id,
			ID:     devices.ID{ID: id},
			Online: true,
			State: devices.State{
				"temperature": 0.0,
				"humidity":    0.0,
			},
		}
		node.AddOrUpdate(dev)
	}

	if dataType == C.TELLSTICK_TEMPERATURE {
		t, _ := strconv.ParseFloat(C.GoString(value), 64)
		logrus.Debugf("Temperature %s : %f\n", id, t)
		if dev.State["temperature"] != t {
			dev.State["temperature"] = t
		}
	} else if dataType == C.TELLSTICK_HUMIDITY {
		h, _ := strconv.ParseFloat(C.GoString(value), 64)
		logrus.Debugf("Humidity %s : %f\n", id, h)
		if dev.State["humidity"] != h {
			dev.State["humidity"] = h
		}
	}

	node.SyncDevice(id)
}

//export deviceEvent
func deviceEvent(deviceID, method int, data *C.char, callbackID int, context unsafe.Pointer) {
	//log.Println("DeviceEVENT: ", deviceID, method, C.GoString(data))
	id := strconv.Itoa(deviceID)
	dev := node.GetDevice(id)
	if dev == nil {
		logrus.Errorf("Unknown device %s", id)
		return
	}
	if method&C.TELLSTICK_TURNON != 0 {
		dev.State["on"] = true
	}
	if method&C.TELLSTICK_TURNOFF != 0 {
		dev.State["on"] = false
	}
	if method&C.TELLSTICK_DIM != 0 {
		level, err := strconv.ParseUint(C.GoString(data), 10, 16)
		if err != nil {
			logrus.Error(err)
			return
		}
		if level == 0 {
			dev.State["on"] = false
		}
		if level > 0 {
			dev.State["on"] = true
		}
		dev.State["brightness"] = float64(level) / 255.0
	}

	node.SyncDevice(id)
}

//export deviceChangeEvent
func deviceChangeEvent(deviceID, changeEvent, changeType, callbackID int, context unsafe.Pointer) {
	//log.Println("DeviceChangeEVENT: ", deviceID, changeEvent, changeType)
}

//export rawDeviceEvent
func rawDeviceEvent(data *C.char, controllerID, callbackID int, context unsafe.Pointer) {
	//log.Println("rawDeviceEVENT: ", controllerID, C.GoString(data))
}
