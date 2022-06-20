package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/jonaz/gombus"
	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/models/devices"
	"github.com/stampzilla/stampzilla-go/v2/pkg/node"
)

var configUpdated chan struct{}
var worker = NewWorker()
var debug bool

func main() {
	if os.Getenv("STAMPZILLA_DEBUG") != "" {
		debug = true
	}
	start()
}

func start() {
	config := NewConfig()

	if debug {
		file, err := os.Open("config.json")
		if err != nil {
			logrus.Error(err)
			return
		}

		err = json.NewDecoder(file).Decode(&config)
		if err != nil {
			logrus.Error(err)
			return
		}
		spew.Dump("config: ", config)
		node := setupNode(config)
		worker.Start(context.Background(), 1)
		startLoops(context.Background(), config, node)
		select {}
	}

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
	mainCtx, mainCancel := context.WithCancel(context.Background())
	worker.Start(mainCtx, 1)

	ctx, cancel := context.WithCancel(mainCtx)
	startLoops(ctx, config, node)
	for {
		select {
		case <-node.Stopped():
			log.Println("Stopping mbus node")
			mainCancel()
			cancel()
			return
		case <-configUpdated:
			cancel()
			ctx, cancel = context.WithCancel(context.Background())
			startLoops(ctx, config, node)
		}
	}
}

func setupNode(config *Config) *node.Node {
	node := node.New("mbus")
	node.OnConfig(updatedConfig(config))
	return node
}

func startLoops(ctx context.Context, config *Config, node *node.Node) {
	for _, d := range config.Devices {
		id := strconv.Itoa(d.PrimaryAddress)
		dev := &devices.Device{
			Name:   d.Name,
			Type:   "sensor",
			ID:     devices.ID{ID: id},
			Online: true,
			Traits: []string{},
			State:  make(devices.State),
		}

		if !debug {
			node.AddOrUpdate(dev)
		}
		go tickerLoop(ctx, config, node, d)
	}
}

func tickerLoop(ctx context.Context, config *Config, node *node.Node, mbusDevice Device) {
	dur, err := time.ParseDuration(mbusDevice.Interval)
	if err != nil {
		logrus.Errorf("wrong duration format %s: %s", config.Interval, err)
		return
	}
	ticker := time.NewTicker(dur)
	logrus.Infof("Config OK. starting fetch loop for %s", dur)
	for {
		select {
		case <-ticker.C:
			// this makes sure we only use the mbus connection 1 at a time.
			worker.Do(func() error {
				newState, err := fetchState(config, mbusDevice)
				if err != nil {
					return fmt.Errorf("error fetching mbus data from device %d: %w", mbusDevice.PrimaryAddress, err)
				}
				if debug {
					spew.Dump("state", newState)
				} else {
					node.UpdateState(strconv.Itoa(mbusDevice.PrimaryAddress), newState)
				}
				return nil
			}, nil)
		case <-ctx.Done():
			// just stop and we will be restarted
			return
		case <-node.Stopped():
			ticker.Stop()
			log.Println("Stopping mbus node")
			return
		}
	}
}

func updatedConfig(config *Config) func(data json.RawMessage) error {
	return func(data json.RawMessage) error {
		err := json.Unmarshal(data, config)

		select {
		case configUpdated <- struct{}{}:
		default:
		}

		return err
	}
}

func fetchState(config *Config, device Device) (devices.State, error) {
	conn, err := gombus.Dial(net.JoinHostPort(config.Host, config.Port))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	_, err = conn.Write(gombus.SndNKE(uint8(device.PrimaryAddress)))
	if err != nil {
		return nil, err
	}
	err = conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	if err != nil {
		return nil, err
	}
	_, err = gombus.ReadSingleCharFrame(conn)
	if err != nil {
		return nil, err
	}

	frame, err := gombus.ReadSingleFrame(conn, device.PrimaryAddress)
	if err != nil {
		return nil, err
	}

	state := make(devices.State)

	if len(device.Frames) > 1 {
		return nil, fmt.Errorf("currently only supports single frame for now")
	}

	for _, f := range device.Frames {
		for _, r := range f {
			dataRecord := frame.DataRecords[r.Id]
			key := fmt.Sprintf("%s_%s", r.Name, dataRecord.Unit.Unit)
			if dataRecord.ValueString != "" {
				state[key] = dataRecord.ValueString
			} else {
				state[key] = dataRecord.Value
			}
		}
	}
	return state, nil
}
