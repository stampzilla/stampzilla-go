package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strconv"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/gozwave"
	"github.com/stampzilla/gozwave/events"
	"github.com/stampzilla/gozwave/nodes"
	zp "github.com/stampzilla/gozwave/protocol"
	"github.com/stampzilla/gozwave/serialrecorder"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models/devices"
	"github.com/stampzilla/stampzilla-go/pkg/node"
)

func main() {
	node := node.New("zwave")

	config := &Config{}
	node.OnConfig(updatedConfig(config))
	wait := node.WaitForFirstConfig()

	err := node.Connect()
	if err != nil {
		logrus.Error(err)
		return
	}

	logrus.Info("Waiting for config from server")
	err = wait()
	if err != nil {
		logrus.Error(err)
		return
	}

	var z *gozwave.Controller
	z, f, err := getZwaveController(config)
	if err != nil {
		logrus.Error(err)
		return
	}
	if f != nil {
		defer f.Close()
	}

	node.OnRequestStateChange(onRequestStateChange(z))

	// Add all existing nodes to the state / device list
	for _, znode := range z.Nodes.All() {
		if znode.Id == 1 {
			continue
		}
		addOrUpdateDevice(node, znode)
	}

	// Listen from events from the zwave-controller
	for {
		select {
		case <-node.Stopped():
			logrus.Info("Shutting down")
			return
		case event := <-z.GetNextEvent():
			logrus.Debugf("zwave event: %#v", event)
			switch e := event.(type) {
			case events.NodeDiscoverd:
				znode := z.Nodes.Get(e.Address)
				if znode != nil {
					addOrUpdateDevice(node, znode) // Device management
				}

			case events.NodeUpdated:
				znode := z.Nodes.Get(e.Address)
				if znode != nil {
					addOrUpdateDevice(node, znode) // Device management
				}
			}
		}
	}
}

func updatedConfig(config *Config) func(data json.RawMessage) error {
	return func(data json.RawMessage) error {
		logrus.Debug("Received config from server:", string(data))
		return json.Unmarshal(data, &config)
	}
}

func addOrUpdateDevice(node *node.Node, znode *nodes.Node) {
	for i := 0; i < len(znode.Endpoints); i++ {
		devid := strconv.Itoa(int(znode.Id) + (i * 1000))

		stateBool := znode.StateBool
		stateFloat := znode.StateFloat
		if i > 0 {
			ep := znode.Endpoint(i)
			stateBool = ep.StateBool
			stateFloat = ep.StateFloat
		}

		newState := devices.State{}
		if v, ok := stateBool["on"]; ok {
			newState["on"] = v
		}
		if v, ok := stateFloat["level"]; ok {
			newState["brightness"] = v / 100
		}
		if v, ok := stateFloat["power_w"]; ok {
			newState["power_w"] = v
		}
		//Dont add if it already exists
		if dev := node.GetDevice(devid); dev != nil {
			if diff := dev.State.Diff(newState); len(diff) != 0 {
				dev.Lock()
				dev.State.MergeWith(diff)
				dev.Unlock()
				node.SyncDevice(devid)
			}
			return
		}

		switch {
		case znode.IsDeviceClass(zp.GENERIC_TYPE_SWITCH_MULTILEVEL,
			zp.SPECIFIC_TYPE_POWER_SWITCH_MULTILEVEL):
			//znode.HasCommand(commands.SwitchMultilevel):
			node.AddOrUpdate(&devices.Device{
				Type:   "light",
				Name:   znode.Brand + " - " + znode.Product + " (Address: " + devid + ")",
				ID:     devices.ID{ID: devid},
				Online: true,
				Traits: []string{"OnOff", "Brightness"},
				State:  newState,
			})
		//case znode.HasCommand(commands.SwitchBinary):
		case znode.IsDeviceClass(zp.GENERIC_TYPE_SWITCH_BINARY,
			zp.SPECIFIC_TYPE_POWER_SWITCH_BINARY):
			node.AddOrUpdate(&devices.Device{
				Type:   "light",
				Name:   znode.Brand + " - " + znode.Product + " (Address: " + devid + ")",
				ID:     devices.ID{ID: devid},
				Online: true,
				Traits: []string{"OnOff"},
				State:  newState,
			})
		}
	}
}

func onRequestStateChange(z *gozwave.Controller) func(state devices.State, device *devices.Device) error {
	return func(state devices.State, device *devices.Device) error {
		logrus.Debugf("OnRequestStateChange for device %s: %#v", device.ID.ID, state)

		id, err := strconv.Atoi(device.ID.ID)
		if err != nil {
			return err
		}
		endpoint := int(id / 1000)
		id = id - (endpoint * 1000)

		eDev := z.Nodes.Get(id)

		if eDev == nil {
			return fmt.Errorf("device %s not found in our config", device.ID.ID)
		}

		var control gozwave.Controllable
		if id < 1000 && len(eDev.Endpoints) < 2 {
			control = eDev
		} else {
			control = eDev.Endpoint(endpoint)
		}

		state.Bool("on", func(on bool) {
			if on {
				logrus.Debugf("turning on %s \n", device.ID.String())
				control.On()
				return
			}
			logrus.Debugf("turning off %s \n", device.ID.String())
			control.Off()
		})

		state.Float("brightness", func(v float64) {
			bri := math.Round(100 * v)
			logrus.Debugf("dimming %s to: %f\n", device.ID.String(), bri)
			eDev.Level(bri)
		})
		return nil
	}
}

func getZwaveController(config *Config) (z *gozwave.Controller, f *os.File, err error) {
	if config.RecordToFile == "" {
		z, err = gozwave.Connect(config.Device, "zwave-networkmap.json")
		return
	}

	f, err = os.Create("/tmp/dat2")
	if err != nil {
		logrus.Error(err)
		return
	}

	re := serialrecorder.New(config.Device, 115200)
	re.Logger = f
	z, err = gozwave.ConnectWithCustomPortOpener(config.Device, "zwave-networkmap.json", re)
	return
}
