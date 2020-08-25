package googleassistant

import (
	"encoding/json"

	"github.com/sirupsen/logrus"
)

// Intent of the action
type Intent string

const (
	// SyncIntent is used when syncing devices
	SyncIntent = "action.devices.SYNC"
	// QueryIntent is used when querying for status
	QueryIntent = "action.devices.QUERY"
	// ExecuteIntent is used when controling devices
	ExecuteIntent = "action.devices.EXECUTE"

	CommandBrightnessAbsolute = "action.devices.commands.BrightnessAbsolute"
	CommandOnOff              = "action.devices.commands.OnOff"
	CommandColorAbsolute      = "action.devices.commands.ColorAbsolute"
)

// Inputs from google
type Inputs []map[string]json.RawMessage

func (i Inputs) Intent() Intent {
	for _, v := range i {
		if v, ok := v["intent"]; ok {
			in := ""
			err := json.Unmarshal(v, &in)
			if err != nil {
				logrus.Error(err)
				return ""
			}

			return Intent(in)
		}
	}
	return ""
}
func (i Inputs) Payload() Payload {
	for _, v := range i {
		if v, ok := v["payload"]; ok {
			pl := Payload{}
			err := json.Unmarshal(v, &pl)
			if err != nil {
				logrus.Error(err)
				return Payload{}
			}
			return pl
		}
	}
	return Payload{}
}

type Payload struct {
	Commands []struct {
		Devices []struct {
			ID string `json:"id"`
			//CustomData struct {
			//FooValue int    `json:"fooValue"`
			//BarValue bool   `json:"barValue"`
			//BazValue string `json:"bazValue"`
			//} `json:"customData"`
		} `json:"devices"`
		Execution []struct {
			Command string `json:"command"`
			Params  struct {
				On         bool `json:"on"`
				Brightness int  `json:"brightness"`
				Color      struct {
					Name        string `json:"name"`
					SpectrumRGB int    `json:"spectrumRGB"`
					Temperature int    `json:"temperature"`
				} `json:"color"`
			} `json:"params"`
		} `json:"execution"`
	} `json:"commands"`
	Devices []struct {
		ID string `json:"id"`
		//CustomData struct {
		//FooValue int    `json:"fooValue"`
		//BarValue bool   `json:"barValue"`
		//BazValue string `json:"bazValue"`
		//} `json:"customData"`
	} `json:"devices"`
}

type Request struct {
	RequestID string
	Inputs    Inputs
}
