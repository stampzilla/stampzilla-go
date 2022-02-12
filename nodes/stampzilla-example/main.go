package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/models/devices"
	"github.com/stampzilla/stampzilla-go/v2/pkg/node"
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
			"on":         false,
			"brightness": 0.0,
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
	dev4 := &devices.Device{
		Name:   "Device4 that requires a node config",
		Type:   "light",
		ID:     devices.ID{ID: "4"},
		Online: true,
		Traits: []string{"OnOff"},
		State: devices.State{
			"on": false,
		},
	}
	dev5 := &devices.Device{
		Name:   "toggleonoff",
		Type:   "light",
		ID:     devices.ID{ID: "5"},
		Online: true,
		Traits: []string{"OnOff"},
		State: devices.State{
			"on": false,
		},
	}
	dev6 := &devices.Device{
		Name:   "heatpump",
		Type:   "sensor",
		ID:     devices.ID{ID: "6"},
		Online: true,
		Traits: []string{"TemperatureControl"},
		State: devices.State{
			"temperature": 22.0,
		},
	}

	node.OnRequestStateChange(func(state devices.State, device *devices.Device) error {
		logrus.Info("OnRequestStateChange:", state, device.ID)

		// Make device 3 follow the value of device 1
		if device.ID.ID == "1" {
			dev3.State["on"] = state["on"]
			node.AddOrUpdate(dev3)
		}

		// Load device config from the config struct
		devConfig, ok := config.Devices[device.ID.ID]

		// Require a device config for node 4 only
		if !ok && device.ID.ID == "4" {
			return fmt.Errorf("Found no config for device %s", device.ID)
		}

		state.Bool("on", func(on bool) {
			if on {
				fmt.Printf("turning on %s with senderid %s\n", device.ID.String(), devConfig.SenderID)
				return
			}
			fmt.Printf("turning off %s with senderid %s\n", device.ID.String(), devConfig.SenderID)
		})

		state.Float("brightness", func(lvl float64) {
			fmt.Printf("dimming to %f on device %s\n", lvl, device.ID.String())
		})

		return nil
	})

	err := node.Connect()
	if err != nil {
		logrus.Error(err)
		return
	}

	ticker := time.NewTicker(time.Second * 10)
	go func() {
		for range ticker.C {
			dev5.Lock()
			dev5.State.Bool("on", func(on bool) {
				dev5.State["on"] = !on
			})
			dev5.Unlock()
			node.SyncDevice(dev5.ID.ID)
		}
	}()

	node.AddOrUpdate(dev1)
	node.AddOrUpdate(dev2)
	node.AddOrUpdate(dev3)
	node.AddOrUpdate(dev4)
	node.AddOrUpdate(dev5)
	node.AddOrUpdate(dev6)

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
