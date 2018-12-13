package main

import (
	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/models"
	"github.com/stampzilla/stampzilla-go/pkg/node"
)

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
		Traits: traits,
		State: models.DeviceState{
			"on": false,
		},
	}

	if light.StateSwitch != "" {
		tunnel.AddLink(light.StateSwitch, "on", "bool", dev)
	}

	if light.StateBrightness != "" {
		tunnel.AddLink(light.StateBrightness, "brightness", "level", dev)
	}

	err := node.AddOrUpdate(dev)
	if err != nil {
		logrus.Error(err)
	}
}
func setupSensor(node *node.Node, tunnel *tunnel, sensor sensor) {
	dev := &models.Device{
		Name:  sensor.ID,
		ID:    "sensor." + sensor.ID,
		State: make(models.DeviceState),
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
