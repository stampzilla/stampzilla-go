package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models/devices"
	"github.com/stampzilla/stampzilla-go/pkg/node"
)

func main() {
	node := node.New("knx")

	tunnel := newTunnel(node)
	tunnel.OnConnect = func() {
		for _, dev := range node.Devices.All() {
			dev.SetOnline(true)
		}
		node.SyncDevices()
	}
	tunnel.OnDisconnect = func() {
		for _, dev := range node.Devices.All() {
			dev.SetOnline(false)
		}
		node.SyncDevices()
	}

	config := &config{}

	node.OnConfig(updatedConfig(node, tunnel, config))
	node.OnRequestStateChange(func(state devices.State, device *devices.Device) error {
		id := strings.SplitN(device.ID.ID, ".", 2)

		switch id[0] {
		case "light":

			light := config.GetLight(id[1])
			state.Bool("on", func(on bool) {
				err := light.Switch(tunnel, on)
				if err != nil {
					logrus.Error()
				}
			})
			state.Float("brightness", func(v float64) {
				err := light.Brightness(tunnel, v*100)
				if err != nil {
					logrus.Error()
				}
			})

		default:
			return fmt.Errorf("Unknown device type \"%s\"", id[0])
		}
		return nil
	})

	ctx, shutdown := context.WithCancel(context.Background())
	node.OnShutdown(func() {
		shutdown()
	})

	err := node.Connect()
	if err != nil {
		logrus.Error(err)
		return
	}

	tunnel.Start(ctx)

	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors: true,
	})
	logrus.SetReportCaller(false)

	tunnel.Wait()
	node.Wait()
}

func updatedConfig(node *node.Node, tunnel *tunnel, config *config) func(data json.RawMessage) error {
	return func(data json.RawMessage) error {
		config.Lock()
		defer config.Unlock()
		err := json.Unmarshal(data, config)
		if err != nil {
			return err
		}

		tunnel.SetAddress(config.Gateway.Address)

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
