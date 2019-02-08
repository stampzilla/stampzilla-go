package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models/devices"
	"github.com/stampzilla/stampzilla-go/pkg/node"
)

type Config struct {
	Display string                  `json:"display"`
	Players map[string]PlayerConfig `json:"players"`
}

var config Config

func main() {
	node := node.New("linux")

	node.OnConfig(updatedConfig)
	node.OnRequestStateChange(func(state devices.State, device *devices.Device) error {
		logrus.Info("OnRequestStateChange:", state, device.ID)

		switch device.ID.ID {
		case "monitor":
			if state["on"] == true {
				cmd := exec.Command("xset", "dpms", "force", "on")
				//cmd.Env = append(os.Environ(), "DISPLAY=:0")
				_, err := cmd.Output()
				if err != nil {
					return err
				}
			}
			if state["on"] == false {
				cmd := exec.Command("xset", "dpms", "force", "off")
				//cmd.Env = append(os.Environ(), "DISPLAY=:0")
				_, err := cmd.Output()
				if err != nil {
					return err
				}
			}
		}

		return nil
	})

	err := node.Connect()
	if err != nil {
		logrus.Error(err)
		return
	}

	go monitorDpms(node)
	go monitorHealth(node)
	go startPlayers()

	node.Wait()
}

func updatedConfig(data json.RawMessage) error {
	newConf := Config{}
	err := json.Unmarshal(data, &newConf)
	if err != nil {
		return err
	}

	config = newConf

	go restartPlayers()
	return nil
}

func monitorDpms(node *node.Node) {
	dev := &devices.Device{
		Name:   "Monitor",
		ID:     devices.ID{ID: "monitor"},
		Online: true,
		Traits: []string{"OnOff"},
		State: devices.State{
			"on": false,
		},
	}

	re := regexp.MustCompile("Monitor is (in )?([^ \n]+)")

	for {
		cmd := exec.Command("xset", "q")
		//cmd.Env = append(os.Environ(), "DISPLAY=:0")
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		out, err := cmd.Output()

		if err != nil {
			logrus.Errorf("Failed to read monitor status: %s: %s", fmt.Sprint(err), stderr.String())
			return
		}

		status := re.FindStringSubmatch(string(out))
		if len(status) > 2 {
			dev.State["monitor_status"] = status[2]
			dev.State["on"] = status[2] == "On"
			node.AddOrUpdate(dev)
		}
		<-time.After(time.Second * 1)
	}
}

func monitorHealth(node *node.Node) {
	dev := &devices.Device{
		Name:   "Health",
		ID:     devices.ID{ID: "health"},
		Online: true,
		State:  devices.State{},
	}

	for {
		cmd := exec.Command("uptime")
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		out, err := cmd.Output()

		if err != nil {
			logrus.Errorf("Failed to read health: %s: %s", fmt.Sprint(err), stderr.String())
			return
		}

		dev.State["uptime"] = string(out)

		//out, err = exec.Command("free -h | awk -F ' ' 'NR>1 {print $3}'").Output()
		//dev.State["ram"] = string(out)
		//logrus.Warn(err)

		node.AddOrUpdate(dev)
		<-time.After(time.Second * 1)
	}
}
