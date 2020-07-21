package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models/devices"
	"github.com/stampzilla/stampzilla-go/pkg/node"
)

var httpClient = &http.Client{
	Timeout: time.Second * 5,
}

func main() {
	node := node.New("husdata-h60")
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

	//connection := basenode.Connect()
	dev := &devices.Device{
		Name:   "heatpump",
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
		// handle error
		return nil, err
	}
	defer resp.Body.Close()

	hp := &HeatPump{}
	err = json.NewDecoder(resp.Body).Decode(hp)
	return hp, err
}
