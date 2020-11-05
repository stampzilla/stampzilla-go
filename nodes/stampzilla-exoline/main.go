package main

import (
	"bufio"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/sirupsen/logrus"
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
			log.Println("Stopping exoline node")
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
		log.Fatalf("Failed to dial: %v", err)
	}
	defer conn.Close()
	buf := bufio.NewReader(conn)

	state := make(devices.State)
	for _, sensor := range config.Variables {
		decoded, err := hex.DecodeString(sensor.Address)
		if err != nil {
			return nil, err
		}

		switch sensor.Type {
		case "float":
			addr := []byte{0xff, 0x1e, 0xc8, 0x04, 0xb6, 0x04, 0x08, 0x00}
			addr[5] = decoded[0]
			addr[6] = decoded[1]
			addr[7] = decoded[2]
			data, err := Send(buf, conn, addr)
			if err != nil {
				return nil, err
			}

			if data[1] != 0x05 && data[2] != 0x00 {
				return nil, nil
			}

			f, err := asRoundedFloat(data[3 : len(data)-2])
			if err != nil {
				return nil, fmt.Errorf("data %s %w", printHex(data), err)
			}

			logrus.Debug("decode float:", f)

			state[sensor.Name] = f
		}
	}

	return state, nil
}

/*
	fmt.Println("frånluft") // Extract air
	Send(buf, w, []byte{0xff, 0x1e, 0xc8, 0x04, 0xb6, 0x04, 0x08, 0x00})

	fmt.Println("ute") // Outdoor air
	Send(buf, w, []byte{0xff, 0x1e, 0xc8, 0x04, 0xb6, 0x04, 0x04, 0xec})

	fmt.Println("avluft") // Exhaust air
	Send(buf, w, []byte{0xff, 0x1e, 0xc8, 0x04, 0xb6, 0x3b, 0x00, 0x95})

	fmt.Println("tilluft") // Supply air
	Send(buf, w, []byte{0xff, 0x1e, 0xc8, 0x04, 0xb6, 0x39, 0x02, 0xfa})

	fmt.Println("börvärde") // Set temperature
	Send(buf, w, []byte{0xff, 0x1e, 0xc8, 0x04, 0xb6, 0x04, 0x00, 0x08})
*/
