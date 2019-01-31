package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"regexp"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models/devices"
	"github.com/stampzilla/stampzilla-go/pkg/node"
)

func main() {
	node := node.New("linux")

	monitor := &devices.Device{
		Name:   "Monitor",
		ID:     devices.ID{ID: "monitor"},
		Online: true,
		Traits: []string{"OnOff"},
		State: devices.State{
			"on": false,
		},
	}

	health := &devices.Device{
		Name:   "Health",
		ID:     devices.ID{ID: "health"},
		Online: true,
		State:  devices.State{},
	}

	node.OnConfig(updatedConfig)
	node.OnRequestStateChange(func(state devices.State, device *devices.Device) error {
		logrus.Info("OnRequestStateChange:", state, device.ID)

		switch device.ID.ID {
		case "monitor":
			if state["on"] == true {
				cmd := exec.Command("xset", "dpms", "force", "on")
				cmd.Env = append(os.Environ(), "DISPLAY=:0")
				_, err := cmd.Output()
				if err != nil {
					return err
				}
			}
			if state["on"] == false {
				cmd := exec.Command("uptime")
				cmd.Env = append(os.Environ(), "DISPLAY=:0")
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

	go monitorDpms(node, monitor)
	go monitorHealth(node, health)

	node.Wait()
}

func updatedConfig(data json.RawMessage) error {
	logrus.Info("DATA:", string(data))
	return nil
}

func monitorDpms(node *node.Node, dev *devices.Device) {
	node.AddOrUpdate(dev)

	re := regexp.MustCompile("Monitor is (in )?([^ \n]+)")

	for {
		cmd := exec.Command("xset", "q")
		cmd.Env = append(os.Environ(), "DISPLAY=:0")
		out, _ := cmd.Output()

		status := re.FindStringSubmatch(string(out))
		if len(status) > 2 {
			dev.State["monitor_status"] = status[2]
			dev.State["on"] = status[2] == "On"
			node.AddOrUpdate(dev)
		}
		<-time.After(time.Second * 1)
	}
}

func monitorHealth(node *node.Node, dev *devices.Device) {
	node.AddOrUpdate(dev)

	for {
		out, _ := exec.Command("uptime").Output()
		dev.State["uptime"] = string(out)

		//out, err = exec.Command("free -h | awk -F ' ' 'NR>1 {print $3}'").Output()
		//dev.State["ram"] = string(out)
		//logrus.Warn(err)

		node.AddOrUpdate(dev)
		<-time.After(time.Second * 1)
	}
}
