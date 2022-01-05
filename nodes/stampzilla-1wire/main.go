package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-1wire/onewire"
	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/models/devices"
	"github.com/stampzilla/stampzilla-go/v2/pkg/node"
)

func main() {
	start()
}

func start() {
	config := NewConfig()

	node := setupNode(config)

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

	err = tickerLoop(config, node)
	if err != nil {
		logrus.Error(err)
	}
}

func setupNode(config *Config) *node.Node {
	node := node.New("1wire")
	node.OnConfig(updatedConfig(config))
	return node
}

func updatedConfig(config *Config) func(data json.RawMessage) error {
	return func(data json.RawMessage) error {
		return json.Unmarshal(data, config)
	}
}

func tickerLoop(config *Config, node *node.Node) error {
	sensors, err := onewire.SensorsWithTemperature()
	if err != nil {
		return err
	}

	for _, s := range sensors {
		dev := &devices.Device{
			Name:   s,
			Type:   "sensor",
			ID:     devices.ID{ID: s},
			Online: true,
			State:  make(devices.State),
		}
		temp, err := onewire.Temperature(s)
		if err != nil {
			logrus.Errorf("error fetching temp from sensor %s: %s", s, err)
		} else {
			dev.State["temperature"] = temp
		}

		node.AddOrUpdate(dev)
	}

	dur, err := time.ParseDuration(config.Interval)
	if err != nil {
		return fmt.Errorf("wrong duration format %s: %w", config.Interval, err)
	}
	ticker := time.NewTicker(dur)
	logrus.Infof("Config OK. starting fetch loop for %s", dur)
	for {
		select {
		case <-ticker.C:
			for _, s := range sensors {
				temp, err := onewire.Temperature(s)
				if err != nil {
					logrus.Errorf("error fetching temp from sensor %s: %s", s, err)
					continue
				}

				state := make(devices.State)
				state["temperature"] = temp
				node.UpdateState(s, state)
			}

		case <-node.Stopped():
			ticker.Stop()
			log.Println("Stopping 1wire node")
			return nil
		}
	}
}
