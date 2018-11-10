package main

import (
	"encoding/json"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/models"
)

func main() {

	client := NewWebsocketClient()
	node := NewNode(client)
	node.Type = "example"

	node.OnConfig(updatedConfig)

	node.OnRequestStateChange(func(state models.DeviceState, device *models.Device) error {
		logrus.Info("OnRequestStateChange:", state, device.ID)
		return nil
	})

	err := node.Connect()

	if err != nil {
		logrus.Error(err)
		return
	}

	dev1 := &models.Device{
		Name:   "Device1",
		ID:     "1",
		Online: true,
		Traits: []string{"OnOff"},
		State: models.DeviceState{
			"on": false,
		},
	}
	err = node.AddOrUpdate(dev1)
	if err != nil {
		logrus.Error(err)
	}

	dev2 := &models.Device{
		Name:   "Device2",
		ID:     "2",
		Online: true,
		Traits: []string{"OnOff", "Brightness", "ColorSetting"},
		State: models.DeviceState{
			"on": false,
		},
	}
	err = node.AddOrUpdate(dev2)
	if err != nil {
		logrus.Error(err)
	}
	dev3 := &models.Device{
		Name:   "Device3",
		ID:     "3",
		Online: true,
		State: models.DeviceState{
			"on": false,
		},
	}
	err = node.AddOrUpdate(dev3)
	if err != nil {
		logrus.Error(err)
	}

	node.Wait()
	node.Client.Wait()
}

func updatedConfig(data json.RawMessage) error {
	logrus.Info("DATA:", string(data))
	return nil
}
