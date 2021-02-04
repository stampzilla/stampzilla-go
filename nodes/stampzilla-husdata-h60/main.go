package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/models/devices"
	"github.com/stampzilla/stampzilla-go/v2/pkg/node"
)

var httpClient = &http.Client{
	Timeout: time.Second * 5,
}

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
	node := node.New("husdata-h60")
	node.OnConfig(updatedConfig(config))

	node.OnRequestStateChange(func(state devices.State, device *devices.Device) error {
		var err error
		state.Float("RoomTempSetpoint", func(v float64) {
			err = update(config, "0203", v)
		})
		return err
	})

	return node
}

func tickerLoop(config *Config, node *node.Node) {
	dev := &devices.Device{
		Name:   "heatpump",
		Type:   "sensor",
		ID:     devices.ID{ID: "1"},
		Online: true,
		Traits: []string{"TemperatureControl"},
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
	for {
		select {
		case <-ticker.C:
			heatPump, err := fetch(config)
			if err != nil {
				logrus.Errorf("error fetching heatPump data: %s", err)
				continue
			}

			if heatPump.RadiatorForward != 0 { // handle when we get invalid data sometimes. Dont log that
				node.UpdateState(dev.ID.ID, heatPump.State())
			}
		case <-node.Stopped():
			ticker.Stop()
			log.Println("Stopping husdata-h60 node")
			return
		}
	}
}

func updatedConfig(config *Config) func(data json.RawMessage) error {
	return func(data json.RawMessage) error {
		return json.Unmarshal(data, config)
	}
}

func fetch(config *Config) (*HeatPump, error) {
	u, err := url.Parse(config.Host)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, "api", "alldata")
	resp, err := httpClient.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	hp := &HeatPump{}
	err = json.NewDecoder(resp.Body).Decode(hp)
	return hp, err
}

func update(config *Config, id string, value float64) error {
	u, err := url.Parse(config.Host)
	if err != nil {
		return err
	}

	u.Path = path.Join(u.Path, "api", "set")
	q := u.Query()
	q.Set("idx", id)
	q.Set("val", strconv.Itoa(int(math.RoundToEven(value*float64(10.0)))))
	u.RawQuery = q.Encode()

	logrus.Debugf("requesting update value in heatpump %s", u.String())

	resp, err := httpClient.Get(u.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("wrong status code. expected 200, got %d", resp.StatusCode)
	}

	return nil
}
