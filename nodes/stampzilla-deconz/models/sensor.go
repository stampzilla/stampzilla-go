package models

import (
	"strings"

	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models/devices"
)

/*
{
    "config": {
      "battery": 100,
      "offset": 0,
      "on": true,
      "reachable": true
    },
    "ep": 1,
    "etag": "1f0d135e4c7bc074090166ea9fbac900",
    "manufacturername": "LUMI",
    "modelid": "lumi.sensor_ht",
    "name": "Humidity 4",
    "state": {
      "humidity": 2056,
      "lastupdated": "2019-01-20T20:14:18"
    },
    "type": "ZHAHumidity",
    "uniqueid": "00:15:8d:00:02:3d:26:5e-01-0405"
  }
*/
// Lights is a list of Light's
type Sensors map[string]Sensor

func NewSensors() Sensors {
	return make(Sensors)
}

type Sensor struct {
	Config           devices.State `json:"config"`
	Ep               int           `json:"ep"`
	Manufacturername string        `json:"manufacturername"`
	Modelid          string        `json:"modelid"`
	Name             string        `json:"name"`
	State            devices.State `json:"state"`
	Type             string        `json:"type"`
	UniqueID         string        `json:"uniqueid"`
	ID               string        `json:"id"`
}

//Get ID returns different IDs if its a light or sensor. This is because a single sensor device can be devided in multiple sensors in the API
func (s *Sensor) GetID() string {
	// take first part of uniqueid 00:15:8d:00:02:55:82:0f-01-0402 which is the mac address
	return strings.SplitN(s.UniqueID, "-", 2)[0]
}

func (s Sensor) GenerateDevice() *devices.Device {

	online := false
	s.Config.Bool("reachable", func(v bool) { online = v })
	state := devices.State{}
	SensorToDeviceState(s.State, state)

	s.Config.Float("battery", func(b float64) { state["battery"] = int(b) })

	dev := &devices.Device{
		Type: "sensor",
		ID: devices.ID{
			ID: s.GetID(),
		},
		Name:   s.Name,
		Online: online,
		State:  state,
		//Traits: s.GetTraits(),
	}
	return dev
}
func SensorToDeviceState(sensorsState, state devices.State) bool {
	//for k, v := range sensorsState {
	//fmt.Printf("%s: %T %v\n", k, v, v)
	//}
	changes := 0
	sensorsState.String("lastupdated", func(s string) {
		state["lastupdated"] = s
	})

	sensorsState.Float("temperature", func(t float64) {
		tf := t / 100.0
		if state["temperature"] != tf {
			changes++
		}
		state["temperature"] = tf
	})
	sensorsState.Float("humidity", func(h float64) {
		hf := h / 100.0
		if state["temperature"] != hf {
			changes++
		}
		state["humidity"] = hf
	})

	return changes > 0
}
