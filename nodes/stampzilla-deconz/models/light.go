package models

import (
	"math"

	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/models/devices"
)

// Lights is a list of Light's
type Lights map[string]Light

func NewLights() Lights {
	return make(map[string]Light)
}

// Light is a deconz RESP api light
type Light struct {
	Etag         string `json:"etag"`
	Hascolor     bool   `json:"hascolor"`
	Manufacturer string `json:"manufacturer"`
	Modelid      string `json:"modelid"`
	Name         string `json:"name"`
	State        devices.State
	//Pointsymbol  struct {
	//} `json:"pointsymbol"`
	//State struct {
	//Alert     string    `json:"alert"`
	//Bri       int       `json:"bri"`
	//Colormode string    `json:"colormode"`
	//Ct        int       `json:"ct"`
	//Effect    string    `json:"effect"`
	//Hue       int       `json:"hue"`
	//On        bool      `json:"on"`
	//Reachable bool      `json:"reachable"`
	//Sat       int       `json:"sat"`
	//Xy        []float64 `json:"xy"`
	//} `json:"state"`
	Swversion string `json:"swversion"`
	Type      string `json:"type"`
	Uniqueid  string `json:"uniqueid"`
}

func LightToDeviceState(lightState, state devices.State) bool {

	changes := 0
	lightState.Float("bri", func(f float64) {
		if f != state["brightness"] {
			changes++
		}
		state["brightness"] = f / 255
	})

	lightState.Bool("on", func(v bool) {
		if v != state["on"] {
			changes++
		}
		state["on"] = v
		if !v {
			state["brightness"] = 0.0
		}
	})

	lightState.Float("ct", func(v float64) {
		v = (500 + 153) - v // invert value
		temp := float64(math.Round(((v - 153) / (500 - 153) * (6500 - 2000)) + 2000))
		if temp != state["temperature"] {
			changes++
		}
		state["temperature"] = temp
	})

	return changes > 0
}

func (l Light) GenerateDevice(id string) *devices.Device {

	online := false
	l.State.Bool("reachable", func(v bool) { online = v })
	state := devices.State{}
	LightToDeviceState(l.State, state)

	dev := &devices.Device{
		Type: l.GetType(),
		ID: devices.ID{
			ID: id,
		},
		Name:   l.Name,
		Online: online,
		State:  state,
		Traits: l.GetTraits(),
	}
	return dev
}

// GetTraits converts deconz types to stampzilla devices types
func (l Light) GetTraits() []string {
	var traits []string
	//traits :=
	switch l.Type {
	case "On/Off plug-in unit":
		traits = []string{
			"OnOff",
		}
	case "Color temperature light":
		traits = []string{
			"OnOff",
			"ColorSetting",
			"Brightness",
		}
	}

	return traits
}

// GetType converts deconz types to stampzilla devices types
func (l Light) GetType() string {
	switch l.Type {
	case "On/Off plug-in unit":
		return "switch"
	case "Color temperature light":
		return "light"
	}

	return "light"
}
