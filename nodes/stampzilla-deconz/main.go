package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/models/devices"
	"github.com/stampzilla/stampzilla-go/pkg/node"
)

func main() {
	node := node.New("deconz")

	err := localConfig.Load()
	if err != nil {
		logrus.Error(err)
		return
	}

	api := NewAPI(localConfig.APIKey, config)

	node.OnConfig(updatedConfig(node, api))
	node.OnShutdown(func() {
		err := localConfig.Save()
		if err != nil {
			log.Println(err)
			return
		}
	})

	node.OnRequestStateChange(func(state devices.State, device *devices.Device) error {
		logrus.Info("OnRequestStateChange:", state, device.ID)

		u := fmt.Sprintf("lights/%s/state", device.ID.ID)

		lightState := make(map[string]interface{})
		if b, ok := state["brightness"]; ok {
			bri := int(math.Round(255 * b.(float64)))
			lightState["bri"] = bri
			if bri != 0 {
				lightState["on"] = true
			}
		}
		if b, ok := state["on"]; ok {
			lightState["on"] = b
		}

		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)
		err = enc.Encode(lightState)
		if err != nil {
			return err
		}
		return api.Put(u, &buf, nil)
	})

	err = node.Connect()

	if err != nil {
		logrus.Error(err)
		return
	}

	node.Wait()
}

func updatedConfig(node *node.Node, api *API) func(data json.RawMessage) error {
	return func(data json.RawMessage) error {
		logrus.Info("Received config from server:", string(data))

		newConf := &Config{}
		err := json.Unmarshal(data, newConf)
		if err != nil {
			return err
		}

		// example when we change "global" config
		if newConf.IP != config.IP || newConf.Port != config.Port {
			fmt.Println("ip changed. TODO lets connect to that instead")
		}

		config.IP = newConf.IP
		config.Port = newConf.Port
		config.Password = newConf.Password
		logrus.Info("Config is now: ", config)

		if localConfig.APIKey == "" {
			createUser()
			api.Key = localConfig.APIKey
		}

		syncLights(node, api)

		return nil
	}
}

func syncLights(node *node.Node, api *API) error {

	lights, err := api.Lights()
	if err != nil {
		return err
	}

	foundStateChange := false
	for id, light := range lights {

		dev := node.GetDevice(id)
		if dev == nil {
			newDev := &devices.Device{
				Type: "light",
				ID: devices.ID{
					Node: node.UUID,
					ID:   id,
				},
				Name:   light.Name,
				Online: true, //TODO use state["reashable"] to set here
				State:  light.State,
				Traits: []string{
					"OnOff",
					"Brightness",
					"ColorSetting",
				},
			}
			node.AddOrUpdate(newDev)
			continue
		}

		foundStateChange = true
		dev.State = light.State
	}

	if foundStateChange {
		node.SyncDevices()
	}
	return nil
}
