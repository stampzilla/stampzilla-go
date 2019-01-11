package main

import (
	"encoding/json"
	"fmt"

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

		devConfig, ok := config.Devices[device.ID.ID]
		if !ok {
			return fmt.Errorf("Foudn no config for device %s", device.ID)
		}

		state.Bool("on", func(on bool) {
			if on {
				fmt.Printf("turning on %s with senderid %s\n", device.ID.String(), devConfig.SenderID)
				return
			}
			fmt.Printf("turning off %s with senderid %s\n", device.ID.String(), devConfig.SenderID)
		})

		state.Float("brightness", func(lvl float64) {
			fmt.Printf("dimming to %f on device %s with senderid %s\n", lvl, device.ID.String(), devConfig.SenderID)
		})

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

var config = &Config{}

func updatedConfig(data json.RawMessage) error {
	logrus.Info("Received config from server:", string(data))

	newConf := &Config{}
	err := json.Unmarshal(data, newConf)
	if err != nil {
		return err
	}

	// example when we change "global" config
	if newConf.GatewayIP != config.GatewayIP {
		fmt.Println("ip changed. lets connect to that instead")
	}

	config = newConf
	logrus.Info("Config is now: ", config)

	return nil
}

type Config struct {
	Devices map[string]struct {
		SenderID string
		RecvEEPs []string // example config taken from enocean node
	}
	GatewayIP string
}

/*
Config to put into gui:
{
	"devices":{
		"1":{
			"senderid":"senderid1",
			"recveeps":[
				"asdf1",
				"asdf2"
			]
		},
		"2":{
			"senderid":"senderid1",
			"recveeps":[
				"asdf1",
				"asdf2"
			]
		}
	}
}

*/
