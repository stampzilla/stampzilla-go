package main

import (
	"encoding/json"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models/devices"
	"github.com/stampzilla/stampzilla-go/pkg/node"
)

type Config struct {
	Display string                  `json:"display"`
	Players map[string]PlayerConfig `json:"players"`
}

var config Config
var n *node.Node

func main() {
	n = node.New("linux")

	n.OnConfig(updatedConfig)
	n.OnRequestStateChange(func(state devices.State, device *devices.Device) error {
		logrus.Info("OnRequestStateChange:", state, device.ID)

		devID := strings.Split(device.ID.ID, ":")
		switch devID[0] {
		case "monitor":
			return changeDpmsState(":"+devID[1], state["on"] == true)
		case "player":
			return commandPlayer(devID[1], state["on"] == true)
		case "audio":
			return commandVolume(state)
		}

		return nil
	})

	err := n.Connect()
	if err != nil {
		logrus.Error(err)
		return
	}

	go startMonitorDpms()
	go monitorHealth()
	go monitorVolume()

	n.Wait()
}

func updatedConfig(data json.RawMessage) error {
	newConf := Config{}
	err := json.Unmarshal(data, &newConf)
	if err != nil {
		return err
	}

	config = newConf

	restartPlayers()

	return nil
}
