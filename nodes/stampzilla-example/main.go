package main

import (
	"encoding/json"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/models/devices"
	"github.com/stampzilla/stampzilla-go/pkg/node"
)

func main() {
	node := node.New("example")

	node.OnConfig(updatedConfig)

	dev1 := &devices.Device{
		Name:   "Device1",
		Type:   "light",
		ID:     devices.ID{ID: "1"},
		Online: true,
		Traits: []string{"OnOff"},
		State: devices.State{
			"on": false,
		},
	}
	dev2 := &devices.Device{
		Name:   "Device2",
		Type:   "light",
		ID:     devices.ID{ID: "2"},
		Online: true,
		Traits: []string{"OnOff", "Brightness", "ColorSetting"},
		State: devices.State{
			"on": false,
		},
	}
	dev3 := &devices.Device{
		Name:   "Device3",
		Type:   "light",
		ID:     devices.ID{ID: "3"},
		Online: true,
		State: devices.State{
			"on": false,
		},
	}

	node.OnRequestStateChange(func(state devices.State, device *devices.Device) error {
		logrus.Info("OnRequestStateChange:", state, device.ID)

		if device.ID.ID == "1" {
			dev3.State["on"] = state["on"]
			node.AddOrUpdate(dev3)
		}
		return nil
	})

	err := node.Connect()

	if err != nil {
		logrus.Error(err)
		return
	}

	err = node.AddOrUpdate(dev1)
	if err != nil {
		logrus.Error(err)
	}

	err = node.AddOrUpdate(dev2)
	if err != nil {
		logrus.Error(err)
	}

	err = node.AddOrUpdate(dev3)
	if err != nil {
		logrus.Error(err)
	}

	node.Wait()
}

func updatedConfig(data json.RawMessage) error {
	logrus.Info("Received config from server:", string(data))
	return nil
}
