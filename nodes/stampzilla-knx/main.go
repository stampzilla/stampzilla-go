package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/models/devices"
	"github.com/stampzilla/stampzilla-go/pkg/node"
	"github.com/stampzilla/stampzilla-go/pkg/websocket"
)

func main() {

	client := websocket.New()
	node := node.New(client)
	node.Type = "knx"

	tunnel := newTunnel(node)
	tunnel.OnConnect = func() {
		for _, dev := range node.Devices.All() {
			dev.Online = true
		}
		node.SyncDevices()
	}
	tunnel.OnDisconnect = func() {
		for _, dev := range node.Devices.All() {
			dev.Online = false
		}
		node.SyncDevices()
	}

	config := &config{}

	node.OnConfig(updatedConfig(node, tunnel, config))
	node.OnRequestStateChange(func(state devices.State, device *devices.Device) error {
		id := strings.SplitN(device.ID, ".", 2)

		switch id[0] {
		case "light":
			config.Lock()
			defer config.Unlock()

			for _, light := range config.Lights {
				if light.ID != id[1] {
					continue
				}

				for stateKey, newState := range state {
					switch stateKey {
					case "on":
						diff, value, err := boolDiff(newState, device.State[stateKey])
						if err != nil {
							return err
						}

						if diff {
							err := light.Switch(tunnel, value)
							if err != nil {
								return err
							}
						}
					case "brightness":
						diff, value, err := scalingDiff(newState, device.State[stateKey])
						if err != nil {
							return err
						}

						if diff {
							err := light.Brightness(tunnel, value)
							if err != nil {
								return err
							}
						}
					}

				}
			}
		default:
			return fmt.Errorf("Unknown device type \"%s\"", id[0])
		}
		return nil
	})

	err := node.Connect()
	if err != nil {
		logrus.Error(err)
		return
	}

	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors: true,
	})
	logrus.SetReportCaller(false)

	node.OnShutdown(func() {
		tunnel.Close()
	})

	node.Wait()
	node.Client.Wait()
}

func updatedConfig(node *node.Node, tunnel *tunnel, config *config) func(data json.RawMessage) error {
	return func(data json.RawMessage) error {
		var configString string
		err := json.Unmarshal(data, &configString)
		if err != nil {
			return err
		}

		config.Lock()
		defer config.Unlock()
		err = json.Unmarshal([]byte(configString), config)
		if err != nil {
			return err
		}

		go tunnel.SetAddress(config.Gateway.Address)

		tunnel.ClearAllLinks()
		for _, light := range config.Lights {
			setupLight(node, tunnel, light)
		}

		for _, sensor := range config.Sensors {
			setupSensor(node, tunnel, sensor)
		}

		return nil
	}
}
