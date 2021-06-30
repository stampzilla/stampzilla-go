package main

import (
	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/models/devices"
	"github.com/stampzilla/stampzilla-go/v2/pkg/node"
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
	node.AddOrUpdate(dev)

	if light.StateSwitch != "" {
		tunnel.AddLink(light.StateSwitch, "on", "bool", dev)
	}

	if light.StateBrightness != "" {
		tunnel.AddLink(light.StateBrightness, "brightness", "procentage", dev)
	}
}

func setupBlinds(node *node.Node, tunnel *tunnel, light light) {
	traits := []string{}

	if light.ControlSwitch != "" {
		traits = append(traits, "OpenClose")
	}

	dev := &devices.Device{
		Name:   light.ID,
		ID:     devices.ID{ID: "blind." + light.ID},
		Traits: traits,
		State: devices.State{
			"on": false,
		},
	}
	node.AddOrUpdate(dev)

	if light.StateSwitch != "" {
		tunnel.AddLink(light.StateSwitch, "on", "bool", dev)
	}

	if light.StateBrightness != "" {
		tunnel.AddLink(light.StateBrightness, "brightness", "procentage", dev)
	}
}

func setupSensor(node *node.Node, tunnel *tunnel, sensor sensor) {
	dev := &devices.Device{
		Name:  sensor.ID,
		ID:    devices.ID{ID: "sensor." + sensor.ID},
		State: make(devices.State),
	}
	node.AddOrUpdate(dev)

	if sensor.Temperature != "" {
		tunnel.AddLink(sensor.Temperature, "temperature", "temperature", dev)
	}
	if sensor.Motion != "" {
		tunnel.AddLink(sensor.Motion, "motion", "bool", dev)
	}
	if sensor.MotionTrue != "" {
		tunnel.AddLink(sensor.Motion, "motionTrue", "bool", dev)
	}
	if sensor.Lux != "" {
		tunnel.AddLink(sensor.Lux, "lux", "lux", dev)
	}
	if sensor.Humidity != "" {
		tunnel.AddLink(sensor.Humidity, "humidity", "humidity", dev)
	}
	if sensor.Co2 != "" {
		tunnel.AddLink(sensor.Co2, "co2", "co2", dev)
	}
	if sensor.Voc != "" {
		tunnel.AddLink(sensor.Voc, "voc", "voc", dev)
	}
	if sensor.DewPoint != "" {
		tunnel.AddLink(sensor.DewPoint, "dewpoint", "dewPoint", dev)
	}
}
