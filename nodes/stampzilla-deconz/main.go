package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-deconz/models"
	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/models/devices"
	"github.com/stampzilla/stampzilla-go/v2/pkg/node"
	"github.com/stampzilla/stampzilla-go/v2/pkg/websocket"
)

var wsClient = websocket.New()

func main() {
	nodeInstance := node.New("deconz")

	err := localConfig.Load()
	if err != nil {
		logrus.Error(err)
		return
	}

	api := NewAPI(localConfig.APIKey, config)

	ctx, stopWs := context.WithCancel(context.Background())
	changed := make(chan struct{})

	go configChanged(ctx, changed, nodeInstance, api)
	go reader(ctx, nodeInstance, api)

	nodeInstance.OnConfig(updatedConfig(changed, api))
	nodeInstance.OnShutdown(func() {
		stopWs()
		err = localConfig.Save()
		if err != nil {
			log.Println(err)
			return
		}
	})

	nodeInstance.OnRequestStateChange(func(state devices.State, device *devices.Device) error {
		// fmt.Printf("onstate req %v\n", state)
		lightState := make(devices.State)
		state.Float("brightness", func(v float64) {
			bri := int(math.Round(255 * v))
			lightState["bri"] = bri
			lightState["on"] = bri != 0
			state["on"] = bri != 0
		})

		state.Bool("on", func(on bool) {
			lightState["on"] = on
		})

		state.Float("temperature", func(v float64) {
			// 153 (6500K) to 500 (2000K)
			if v > 6500.0 {
				v = 6500.0
			}
			if v < 2000.0 {
				v = 2000.0
			}
			v = (6500 + 2000) - v                                                   // invert value
			ct := int(math.Round(((v - 2000) / (6500 - 2000) * (500 - 153)) + 153)) // rescale value
			// fmt.Println("temperature: ", v)
			// fmt.Println("setting ct to ", ct)
			lightState["ct"] = ct
		})

		if len(lightState) == 0 {
			return fmt.Errorf("found no known config in state-change request")
		}
		u := fmt.Sprintf("lights/%s/state", device.ID.ID)
		err := api.PutData(u, &lightState)
		if err != nil {
			logrus.Error(err)
		}
		return node.ErrSkipSync
	})

	err = nodeInstance.Connect()

	if err != nil {
		logrus.Error(err)
		return
	}

	nodeInstance.Wait()
}

type WsEvent struct {
	Type     string        `json:"t"`  // event
	Event    string        `json:"e"`  // changed
	Resource string        `json:"r"`  // lights/sensors/groups
	ID       string        `json:"id"` // 1
	UniqueID string        `json:"uniqueid"`
	State    devices.State `json:"state"`
	Config   devices.State `json:"config"`
}

// Get ID returns different IDs if its a light or sensor. This is because a single sensor device can be divided in multiple sensors in the API.
func (we *WsEvent) GetID() string {
	if we.Resource == "sensors" {
		// take first part of uniqueid 00:15:8d:00:02:55:82:0f-01-0402 which is the mac address
		return strings.SplitN(we.UniqueID, "-", 2)[0]
	}

	return we.ID
}

func reader(ctx context.Context, node *node.Node, api *API) {
	for {
		select {
		case <-ctx.Done():
			logrus.Info("Stopping reader because:", ctx.Err())
			return
		case data := <-wsClient.Read():
			event := &WsEvent{}
			err := json.Unmarshal(data, event)
			if err != nil {
				logrus.Error(err)
				continue
			}

			logrus.Tracef("event: %#v\n", event)

			if event.Type != "event" {
				continue
			}

			dev := node.GetDevice(event.GetID())
			if dev == nil {
				switch event.Resource {
				case "sensors":
					err = syncSensors(node, api)
					if err != nil {
						logrus.Error(err)
						continue
					}
				case "lights":
					err = syncLights(node, api)
					if err != nil {
						logrus.Error(err)
						continue
					}
				}
				continue
			}

			newState := make(devices.State)
			if event.Resource == "lights" {
				models.LightToDeviceState(event.State, newState)
			}
			if event.Resource == "sensors" {
				models.SensorToDeviceState(event.State, newState)
			}

			// reachable again
			//{"e":"changed","id":"1","r":"lights","state":{"reachable":true},"t":"event","uniqueid":"00:0b:57:ff:fe:c0:28:82-01"}
			event.State.Bool("reachable", func(online bool) {
				dev.SetOnline(online)
				if online != dev.Online {
					node.SyncDevice(dev.ID.ID)
				}
			})
			// sensors have reachable in Config
			event.Config.Bool("reachable", func(online bool) {
				dev.SetOnline(online)
				if online != dev.Online {
					node.SyncDevice(dev.ID.ID)
				}
			})

			// Sensor have battery in Config
			event.Config.Float("battery", func(b float64) {
				newState["battery"] = int(b)
			})

			node.UpdateState(dev.ID.ID, newState)

			// SENSOR
			//{"config":{"battery":100,"offset":0,"on":true,"reachable":true},"e":"changed","id":"3","r":"sensors","t":"event","uniqueid":"00:15:8d:00:02:3d:26:5e-01-0402"}
			//{"e":"changed","id":"3","r":"sensors","state":{"lastupdated":"2019-01-20T19:15:51","temperature":2581},"t":"event","uniqueid":"00:15:8d:00:02:3d:26:5e-01-0402"}
			//{"config":{"battery":100,"offset":0,"on":true,"reachable":true},"e":"changed","id":"4","r":"sensors","t":"event","uniqueid":"00:15:8d:00:02:3d:26:5e-01-0405"}
			//{"e":"changed","id":"4","r":"sensors","state":{"humidity":2289,"lastupdated":"2019-01-20T19:15:51"},"t":"event","uniqueid":"00:15:8d:00:02:3d:26:5e-01-0405"}

			// TODO
			// {"e":"changed","id":"1","r":"lights","state":{"on":false},"t":"event","uniqueid":"00:0b:57:ff:fe:c0:28:82-01"}
			// http://dresden-elektronik.github.io/deconz-rest-doc/websocket/
		}
	}
}

func configChanged(parentCtx context.Context, changed chan struct{}, node *node.Node, api *API) {
	var err error
	for {
		var ctx context.Context
		var cancel context.CancelFunc
		select {
		case <-parentCtx.Done():
			return
		case <-changed:
			err = syncLights(node, api)
			if err != nil {
				logrus.Error(err)
			}
			err = syncSensors(node, api)
			if err != nil {
				logrus.Error(err)
			}
			if cancel != nil {
				cancel()
			}
			ctx, cancel = context.WithCancel(parentCtx)
			defer cancel()
			wsClient.ConnectWithRetry(ctx, fmt.Sprintf("ws://%s:%s", config.IP, config.WsPort), nil)
		}
	}
}

func updatedConfig(changed chan struct{}, api *API) func(data json.RawMessage) error {
	return func(data json.RawMessage) error {
		logrus.Info("Received config from server:", string(data))

		newConf := &Config{}
		err := json.Unmarshal(data, newConf)
		if err != nil {
			return err
		}

		if newConf.IP == config.IP && newConf.Port == config.Port && newConf.Password == config.Password {
			return nil
		}

		config.IP = newConf.IP
		config.Port = newConf.Port
		config.Password = newConf.Password
		logrus.Info("Config is now: ", config)

		if localConfig.APIKey == "" {
			createUser()
			api.Key = localConfig.APIKey
		}

		config.WsPort = getWsPort(api)
		logrus.Info("Ws port is: ", config.WsPort)

		changed <- struct{}{}
		return nil
	}
}

/*
{
  "1": {
    "ctmax": 454,
    "ctmin": 250,
    "etag": "1926a9a4a0a2f03deafdc9b5749770c1",
    "hascolor": true,
    "manufacturername": "IKEA of Sweden",
    "modelid": "TRADFRI bulb E27 WS opal 980lm",
    "name": "Light 1",
    "state": {
      "alert": "none",
      "bri": 25,
      "colormode": "xy",
      "ct": 359,
      "on": false,
      "reachable": true
    },
    "swversion": "1.2.217",
    "type": "Color temperature light",
    "uniqueid": "00:0b:57:ff:fe:c0:28:82-01"
  },
  "2": {
    "etag": "7a23e35470c7aeae365325a826f610ca",
    "hascolor": false,
    "manufacturername": "IKEA of Sweden",
    "modelid": "TRADFRI control outlet",
    "name": "Light 2",
    "state": {
      "alert": "none",
      "bri": 194,
      "on": false,
      "reachable": true
    },
    "swversion": "2.0.019",
    "type": "On\/Off plug-in unit",
    "uniqueid": "00:0d:6f:ff:fe:ac:33:bd-01"
  }
}

*/

func syncLights(node *node.Node, api *API) error {
	lights, err := api.Lights()
	if err != nil {
		return err
	}

	change := 0
	for id, light := range lights {
		dev := node.GetDevice(id)
		if dev == nil {
			newDev := light.GenerateDevice(id)
			node.AddOrUpdate(newDev)
			continue
		}

		if models.LightToDeviceState(light.State, dev.State) {
			change++
		}
	}

	if change > 0 {
		node.SyncDevices()
	}
	return nil
}

func syncSensors(node *node.Node, api *API) error {
	sensors, err := api.Sensors()
	if err != nil {
		return err
	}

	change := 0
	for _, sensor := range sensors {
		if sensor.Modelid == "PHDL00" {
			continue // Skip default daylight "virtual/fake" sensor
		}
		dev := node.GetDevice(sensor.GetID())
		if dev == nil {
			newDev := sensor.GenerateDevice()
			node.AddOrUpdate(newDev)
			continue
		}

		if models.SensorToDeviceState(sensor.State, dev.State) {
			change++
		}
	}

	if change > 0 {
		node.SyncDevices()
	}
	return nil
}

func getWsPort(api *API) string {
	type config struct {
		Websocketport int
	}

	c := &config{}

	err := api.Get("config", c)
	if err != nil {
		logrus.Error(err)
		return ""
	}

	return strconv.Itoa(c.Websocketport)
}
