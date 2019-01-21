package main

import (
	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/models/devices"
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

	dev := &devices.Device{
		Name:   light.ID,
		ID:     devices.ID{ID: "light." + light.ID},
		Traits: traits,
		State: devices.State{
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
	dev := &devices.Device{
		Name:  sensor.ID,
		ID:    devices.ID{ID: "sensor." + sensor.ID},
		State: make(devices.State),
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
	if sensor.Humidity != "" {
		tunnel.AddLink(sensor.Humidity, "humidity", "humidity", dev)
	}

	err := node.AddOrUpdate(dev)
	if err != nil {
		logrus.Error(err)
	}
}
