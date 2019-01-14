package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strconv"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/models/devices"
	"github.com/stampzilla/stampzilla-go/pkg/node"
	"github.com/stampzilla/stampzilla-go/pkg/websocket"
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
			return nil
		}
		u := fmt.Sprintf("lights/%s/state", device.ID.ID)
		return api.PutData(u, &lightState)
	})

	err = node.Connect()

	if err != nil {
		logrus.Error(err)
		return
	}

	node.Wait()
}

var stopWs context.CancelFunc

func updatedConfig(node *node.Node, api *API) func(data json.RawMessage) error {
	return func(data json.RawMessage) error {
		logrus.Info("Received config from server:", string(data))

		newConf := &Config{}
		err := json.Unmarshal(data, newConf)
		if err != nil {
			return err
		}

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

		config.WsPort = getWsPort(api)
		logrus.Info("Ws port is: ", config.WsPort)

		ws := websocket.New()

		var ctx context.Context
		ctx, stopWs = context.WithCancel(context.Background())
		ws.ConnectWithRetry(ctx, fmt.Sprintf("ws://%s:%s", config.IP, config.WsPort), nil)

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
