package main

import (
	"encoding/json"
	"fmt"
	"time"

	client "github.com/influxdata/influxdb1-client/v2"
	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models/devices"
	"github.com/stampzilla/stampzilla-go/pkg/node"
)

// Config holds the influxdb connection details
type Config struct {
	Host     string
	Port     string
	Username string
	Password string
	Database string
}

var config = &Config{}

var influxClient client.Client

var deviceList = devices.NewList()

func main() {
	node := node.New("metrics-influxdb")

	stop := make(chan struct{})
	device := make(chan func(), 1000)
	node.OnConfig(updatedConfig(stop, device))
	node.On("devices", onDevices(device))
	err := node.Connect()

	if err != nil {
		logrus.Error(err)
		return
	}

	defer func() {
		if influxClient != nil {
			influxClient.Close()
		}
	}()
	node.Wait()
}
func onDevices(deviceChan chan func()) func(data json.RawMessage) error {
	return func(data json.RawMessage) error {
		devs := make(devices.DeviceMap)
		err := json.Unmarshal(data, &devs)
		if err != nil {
			return err
		}

		for _, d := range devs {
			device := d
			deviceChan <- func() {
				//check if state is different
				state := make(devices.State)
				if prevDev := deviceList.Get(device.ID); prevDev != nil {
					state = prevDev.State.Diff(device.State)
					prevDev.State.MergeWith(device.State)
					prevDev.Alias = device.Alias
					prevDev.Name = device.Name
					prevDev.Type = device.Type
				} else {
					if device.State == nil {
						device.State = make(devices.State)
					}
					state = device.State
					deviceList.Add(device)
				}

				if len(state) > 0 {
					logrus.Infof("We should log value node: %s, %s  %#v", device.ID.Node, device.Name, state)
					tags := map[string]string{
						"node-uuid": device.ID.Node,
						"name":      device.Name,
						"alias":     device.Alias,
						"id":        device.ID.ID,
						"type":      device.Type,
					}
					err = write(tags, state)
					if err != nil {
						logrus.Error("error writing to influx: ", err)
					}
				}
			}
		}
		return err
	}
}

func worker(stop chan struct{}, deviceChan chan func()) {
	logrus.Info("Starting worker")
	for {
		select {
		case <-stop:
			logrus.Info("stopping worker")
			return
		case fn := <-deviceChan:
			fn()
		}
	}
}

func updatedConfig(stop chan struct{}, deviceChan chan func()) func(data json.RawMessage) error {
	return func(data json.RawMessage) error {
		logrus.Info("Got new config:", string(data))

		if len(data) == 0 {
			return nil
		}

		err := json.Unmarshal(data, config)
		if err != nil {
			return err
		}

		if config.Database == "" {
			config.Database = "stampzilla"
		}

		if config.Port == "" {
			config.Port = "8086"
		}

		// stop worker if its running
		select {
		case stop <- struct{}{}:
		default:
		}
		influxClient, err = InitClient()
		if err != nil {
			return err
		}

		// start worker
		go worker(stop, deviceChan)

		logrus.Infof("Config is now: %#v", config)
		return nil
	}
}

// InitClient makes a new influx db client
func InitClient() (client.Client, error) {
	return client.NewHTTPClient(client.HTTPConfig{
		Addr:     fmt.Sprintf("http://%s:%s", config.Host, config.Port),
		Username: config.Username,
		Password: config.Password,
	})
}

func write(tags map[string]string, fields map[string]interface{}) error {
	// Create a new point batch
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  config.Database,
		Precision: "s",
	})
	if err != nil {
		return err
	}

	pt, err := client.NewPoint("device", tags, fields, time.Now())
	if err != nil {
		return err
	}
	bp.AddPoint(pt)

	// Write the batch
	return influxClient.Write(bp)
}
