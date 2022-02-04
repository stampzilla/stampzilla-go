package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/models/devices"
	"github.com/stampzilla/stampzilla-go/v2/pkg/node"
)

var httpClient = &http.Client{
	Timeout: time.Second * 20,
}

var pricesStore = NewPrices()

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

	tickerLoop(config, node)
}

func setupNode(config *Config) *node.Node {
	node := node.New("tibber")
	node.OnConfig(updatedConfig(config))

	return node
}

func tickerLoop(config *Config, node *node.Node) {
	dev := &devices.Device{
		Name:   "electricity",
		Type:   "sensor",
		ID:     devices.ID{ID: "1"},
		Online: true,
		// Traits: []string{"SensorState"}, // TODO add SensorState to frontend
		State: make(devices.State),
	}

	node.AddOrUpdate(dev)

	delay := nextDelay()
	timer := time.NewTimer(delay)
	fetchAndCalculate(config, node)
	logrus.Info("scheduling first run in", delay)
	for {
		select {
		case <-timer.C:
			timer.Reset(nextDelay())
			fetchAndCalculate(config, node)

		case <-node.Stopped():
			timer.Stop()
			logrus.Info("stopping tibber node")
			return
		}
	}
}

func updatedConfig(config *Config) func(data json.RawMessage) error {
	return func(data json.RawMessage) error {
		err := json.Unmarshal(data, config)
		if err != nil {
			return err
		}

		dur, err := time.ParseDuration(config.CarChargeDuration)
		if err != nil {
			return fmt.Errorf("wrong duration format %s: %w", config.carChargeDuration, err)
		}
		config.carChargeDuration = dur
		return nil
	}
}

func nextDelay() time.Duration {
	now := time.Now()
	return truncateHour(now).Add(time.Hour).Sub(now)
}

func truncateHour(t time.Time) time.Time {
	t = t.Truncate(time.Minute * 30)
	if t.Minute() > 0 {
		t = t.Add(time.Minute * -1).Truncate(time.Minute * 30)
	}
	return t
}
