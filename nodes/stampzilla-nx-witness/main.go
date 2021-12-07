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
	Timeout: time.Second * 10,
}

func main() {
	if err := start(); err != nil {
		logrus.Error(err)
	}
}

func start() error {
	config := NewConfig()
	nodeInstance := node.New("nx-witness")
	nodeInstance.OnConfig(updatedConfig(config))
	wait := nodeInstance.WaitForFirstConfig()
	api := NewAPI(config)

	nodeInstance.OnRequestStateChange(func(state devices.State, device *devices.Device) error {
		var on bool
		var found bool
		state.Bool("on", func(b bool) {
			on = b
			found = true
		})

		if !found {
			return node.ErrSkipSync
		}

		// Fetch all rules since we need the whole json to make update. API does not support PUT and only changing some fields.
		rules, err := api.fetchEventRules()
		if err != nil {
			return err
		}

		rule, err := rules.ByID(device.ID.ID)
		if err != nil {
			return err
		}
		rule.Disabled = !on
		logrus.Infof("saving rule %#v", rule)

		return api.saveEventRule(rule)
	})

	err := nodeInstance.Connect()
	if err != nil {
		return fmt.Errorf("error connecting: %w", err)
	}

	logrus.Info("Waiting for config from server")
	err = wait()
	if err != nil {
		return err
	}

	tickerLoop(config, nodeInstance, api)

	return nil
}

func tickerLoop(config *Config, node *node.Node, api *API) {
	if err := fetchRules(node, api); err != nil {
		logrus.Error(err)
	}
	timer := time.NewTicker(config.interval)
	logrus.Infof("scheduling first run in %v", config.interval)
	for {
		select {
		case <-timer.C:
			// fetchAndCalculate(config, node)
			err := fetchRules(node, api)
			if err != nil {
				logrus.Error(err)
			}

		case <-node.Stopped():
			timer.Stop()
			logrus.Info("stopping nx-witness node")
			return
		}
	}
}

func fetchRules(node *node.Node, api *API) error {
	rules, err := api.fetchEventRules()
	if err != nil {
		return err
	}

	for _, rule := range rules {
		id := rule.StampzillaDeviceID()
		if id == "" {
			continue
		}

		state := make(devices.State)
		state["on"] = !rule.Disabled

		dev := node.GetDevice(id)
		if dev == nil {
			node.AddOrUpdate(&devices.Device{
				Name:   id,
				Type:   "camera",
				ID:     devices.ID{ID: id},
				Online: true,
				Traits: []string{"OnOff"},
				State:  state,
			})
			continue
		}

		node.UpdateState(id, state)
	}
	return nil
}

func updatedConfig(config *Config) func(data json.RawMessage) error {
	return func(data json.RawMessage) error {
		err := json.Unmarshal(data, config)
		if err != nil {
			return fmt.Errorf("error unmarshal: %w", err)
		}

		dur, err := time.ParseDuration(config.Interval)
		if err != nil {
			return fmt.Errorf("wrong duration format %s: %w", config.interval, err)
		}
		config.interval = dur
		return nil
	}
}
