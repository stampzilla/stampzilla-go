package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strconv"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-deconz/models"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/models/devices"
	"github.com/stampzilla/stampzilla-go/pkg/node"
	nodelib "github.com/stampzilla/stampzilla-go/pkg/node"
	"github.com/stampzilla/stampzilla-go/pkg/websocket"
)

var wsClient = websocket.New()

func main() {
	node := nodelib.New("deconz")

	err := localConfig.Load()
	if err != nil {
		logrus.Error(err)
		return
	}

	api := NewAPI(localConfig.APIKey, config)

	ctx, stopWs := context.WithCancel(context.Background())
	changed := make(chan struct{})

	go configChanged(ctx, changed, node, api)
	go reader(ctx, node)

	node.OnConfig(updatedConfig(changed, api))
	node.OnShutdown(func() {
		stopWs()
		err = localConfig.Save()
		if err != nil {
			log.Println(err)
			return
		}
	})

	node.OnRequestStateChange(func(state devices.State, device *devices.Device) error {

		lightState := make(devices.State)
		state.Float("brightness", func(v float64) {
			bri := int(math.Round(255 * v))
			lightState["bri"] = bri
			lightState["on"] = bri != 0
			state["on"] = lightState["on"]
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
			//fmt.Println("temperature: ", v)
			//fmt.Println("setting ct to ", ct)
			lightState["ct"] = ct
		})

		if len(lightState) == 0 {
			return fmt.Errorf("Found no known config in state-change request")
		}
		u := fmt.Sprintf("lights/%s/state", device.ID.ID)
		err := api.PutData(u, &lightState)
		if err != nil {
			logrus.Error(err)
		}
		return nodelib.ErrSkipSync

	})

	err = node.Connect()

	if err != nil {
		logrus.Error(err)
		return
	}

	node.Wait()
}

type WsEvent struct {
	Type     string        `json:"t"`  // event
	Event    string        `json:"e"`  // changed
	Resource string        `json:"r"`  // lights/sensors/groups
	ID       string        `json:"id"` // 1
	State    devices.State `json:"state"`
}

func reader(ctx context.Context, node *node.Node) {
	for {
		select {
		case <-ctx.Done():
			logrus.Info("Stopping reader because:", ctx.Err())
			return
		case data := <-wsClient.Read():
			event := &WsEvent{}
			json.Unmarshal(data, event)

			if event.Resource != "lights" {
				continue
			}

			dev := node.GetDevice(event.ID)
			if dev == nil {
				continue
			}
			if models.LightToDeviceState(event.State, dev.State) {
				node.SyncDevice(event.ID)
			}

			logrus.Tracef("event: %#v\n", event)
			//TODO
			// {"e":"changed","id":"1","r":"lights","state":{"on":false},"t":"event","uniqueid":"00:0b:57:ff:fe:c0:28:82-01"}
			// http://dresden-elektronik.github.io/deconz-rest-doc/websocket/
		}
	}
}

func configChanged(parentCtx context.Context, changed chan struct{}, node *node.Node, api *API) error {
	for {
		var ctx context.Context
		var cancel context.CancelFunc
		select {
		case <-parentCtx.Done():
			return parentCtx.Err()
		case <-changed:
			syncLights(node, api)
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
