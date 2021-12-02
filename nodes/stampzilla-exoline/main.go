package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-exoline/exoline"
	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/models/devices"
	"github.com/stampzilla/stampzilla-go/v2/pkg/node"
)

func main() {
	node := node.New("exoline")
	config := NewConfig()

	node.OnConfig(updatedConfig(config))
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

	// connection := basenode.Connect()
	dev := &devices.Device{
		Name:   "ventilation",
		Type:   "sensor",
		ID:     devices.ID{ID: "1"},
		Online: true,
		Traits: []string{},
		State:  make(devices.State),
	}

	node.AddOrUpdate(dev)

	dur, err := time.ParseDuration(config.Interval)
	if err != nil {
		logrus.Errorf("wrong duration format %s: %s", config.Interval, err)
		return
	}

	ticker := time.NewTicker(dur)
	logrus.Infof("Config OK. starting fetch loop for %s", dur)

	// first sync once when we start
	sync(config, node, dev)
	for {
		select {
		case <-ticker.C:
			sync(config, node, dev)

		case <-node.Stopped():
			ticker.Stop()
			logrus.Info("Stopping exoline node")
			return
		}
	}
}

func sync(config *Config, node *node.Node, dev *devices.Device) {
	state, err := fetch(config)
	if err != nil {
		logrus.Errorf("error fetching data: %s", err)
		return
	}
	node.UpdateState(dev.ID.ID, state)
}

func updatedConfig(config *Config) func(data json.RawMessage) error {
	return func(data json.RawMessage) error {
		return json.Unmarshal(data, config)
	}
}

func fetch(config *Config) (devices.State, error) {
	var dialer net.Dialer
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	conn, err := dialer.DialContext(ctx, "tcp", config.Host)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %w", err)
	}
	defer conn.Close()
	buf := bufio.NewReader(conn)

	state := make(devices.State)
	for _, sensor := range config.Variables {
		var err error
		var v interface{}
		switch sensor.Type {
		case "float":
			v, err = exoline.RRP(buf, conn, sensor.LoadNumber, sensor.Cell)
			logrus.Debugf("decoded float for %s: %f", sensor.Name, v)
		case "bool":
			v, err = exoline.RLP(buf, conn, sensor.LoadNumber, sensor.Cell)
		case "int":
			v, err = exoline.RXP(buf, conn, sensor.LoadNumber, sensor.Cell)
		}
		if err != nil {
			return nil, fmt.Errorf("error getting %s: %w", sensor.Name, err)
		}
		state[sensor.Name] = v
	}

	return state, nil
}
