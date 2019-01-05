package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"regexp"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/models"
	"github.com/stampzilla/stampzilla-go/pkg/node"
)

func main() {
	node := node.New("linux")

	monitor := &models.Device{
		Name:   "Monitor",
		ID:     "monitor",
		Online: true,
		Traits: []string{"OnOff"},
		State: models.DeviceState{
			"on": false,
		},
	}

	health := &models.Device{
		Name:   "Health",
		ID:     "health",
		Online: true,
		State:  models.DeviceState{},
	}

	node.OnConfig(updatedConfig)
	node.OnRequestStateChange(func(state models.DeviceState, device *models.Device) error {
		logrus.Info("OnRequestStateChange:", state, device.ID)

		switch device.ID {
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

func monitorDpms(node *node.Node, dev *models.Device) {
	err := node.AddOrUpdate(dev)
	if err != nil {
		logrus.Error(err)
	}

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

func monitorHealth(node *node.Node, dev *models.Device) {
	err := node.AddOrUpdate(dev)
	if err != nil {
		logrus.Error(err)
	}

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
