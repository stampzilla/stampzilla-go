package main

import (
	"encoding/json"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/models"
	"github.com/stampzilla/stampzilla-go/pkg/node"
	"github.com/stampzilla/stampzilla-go/pkg/websocket"
)

type config struct {
	Gateway gateway  `json:"gateway"`
	Lights  []light  `json:"lights"`
	Sensors []sensor `json:"sensors"`
}

type gateway struct {
	Address string `json:"address"`
}

type light struct {
	ID                string `json:"id"`
	ControlSwitch     string `json:"control_switch"`
	ControlBrightness string `json:"control_brightness"`
	StateSwitch       string `json:"state_switch"`
	StateBrightness   string `json:"state_brightness"`
}

type sensor struct {
	ID          string `json:"id"`
	Motion      string `json:"motion"`
	Lux         string `json:"lux"`
	Temperature string `json:"temperature"`
}

func main() {

	client := websocket.New()
	node := node.New(client)
	node.Type = "knx"

	tunnel := newTunnel(node)

	node.OnConfig(updatedConfig(node, tunnel))

	node.OnRequestStateChange(func(state models.DeviceState, device *models.Device) error {
		logrus.Info("OnRequestStateChange:", state, device.ID)
		return nil
	})

	err := node.Connect()
	if err != nil {
		logrus.Error(err)
		return
	}

	node.Wait()
	node.Client.Wait()
}

func updatedConfig(node *node.Node, tunnel *tunnel) func(data json.RawMessage) error {
	return func(data json.RawMessage) error {
		var configString string
		err := json.Unmarshal(data, &configString)
		if err != nil {
			return err
		}

		var config config
		err = json.Unmarshal([]byte(configString), &config)
		if err != nil {
			return err
		}

		go tunnel.SetAddress(config.Gateway.Address)

		for _, light := range config.Lights {
			setupLight(node, tunnel, light)
		}

		for _, sensor := range config.Sensors {
			setupSensor(node, tunnel, sensor)
		}

		return nil
	}
}

func setupLight(node *node.Node, tunnel *tunnel, light light) {
	traits := []string{}

	if light.ControlSwitch != "" {
		traits = append(traits, "OnOff")
	}
	if light.ControlBrightness != "" {
		traits = append(traits, "Brightness")
	}

	dev := &models.Device{
		Name:   light.ID,
		ID:     "light." + light.ID,
		Online: true,
		Traits: traits,
		State: models.DeviceState{
			"on": false,
		},
	}
	err := node.AddOrUpdate(dev)
	if err != nil {
		logrus.Error(err)
	}
}
func setupSensor(node *node.Node, tunnel *tunnel, sensor sensor) {
	dev := &models.Device{
		Name:   sensor.ID,
		ID:     "sensor." + sensor.ID,
		Online: true,
		State:  make(models.DeviceState),
	}

	if sensor.Temperature != "" {
		tunnel.AddLink(sensor.Temperature, "temperature", "temperature", dev)
	}
	if sensor.Motion != "" {
		tunnel.AddLink(sensor.Motion, "motion", "bool", dev)
	}
	if sensor.Lux != "" {
		tunnel.AddLink(sensor.Lux, "lux", "lux", dev)
	}

	err := node.AddOrUpdate(dev)
	if err != nil {
		logrus.Error(err)
	}
}
