package main

import (
	"encoding/json"
	"fmt"
	"time"

	client "github.com/influxdata/influxdb1-client/v2"
	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/models"
	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/models/devices"
	"github.com/stampzilla/stampzilla-go/v2/pkg/node"
)

// Config holds the influxdb connection details.
type Config struct {
	Host     string
	Port     string
	Username string
	Password string
	Database string
}

var config = &Config{}

var influxClient client.Client

var (
	nodesList  = make(map[string]string)
	deviceList = devices.NewList()
)

func main() {
	node := node.New("metrics-influxdb")

	stop := make(chan struct{})
	queue := make(chan func(), 1000)
	node.OnConfig(updatedConfig(stop, queue))
	node.On("nodes", onNodes(queue))
	node.On("devices", onDevices(queue))
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

func onNodes(queueChan chan func()) func(data json.RawMessage) error {
	return func(data json.RawMessage) error {
		logrus.Info("Received nodes data")
		nodes := make(map[string]*models.Node)
		err := json.Unmarshal(data, &nodes)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"err": err,
			}).Errorf("failed to unmarshal nodes message")
			return err
		}

		for _, n := range nodes {
			// Only select what we need in order to avoid copying locks on the Node type.
			nodeName := n.Name
			nodeType := n.Type
			nodeUUID := n.UUID
			queueChan <- func() {
				logrus.Infof("update node %s (%s)", nodeUUID, nodeName)
				nodesList[nodeUUID] = nodeType
			}
		}
		return err
	}
}

func onDevices(queueChan chan func()) func(data json.RawMessage) error {
	return func(data json.RawMessage) error {
		devs := make(devices.DeviceMap)
		err := json.Unmarshal(data, &devs)
		if err != nil {
			return err
		}

		for _, d := range devs {
			device := d
			queueChan <- func() {
				// check if state is different
				var state devices.State
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
					tags := map[string]string{
						"node-uuid": device.ID.Node,
						"name":      device.Name,
						"alias":     device.Alias,
						"id":        device.ID.ID,
						"type":      device.Type,
					}

					appendNodeTypeOnTags(tags, device.ID.Node)

					err = write(tags, state)
					if err != nil {
						logrus.WithFields(logrus.Fields{
							"node":   device.ID.Node,
							"device": device.Name,
							"state":  state,
							"err":    err,
						}).Error("error writing to influx")
						return
					}

					logrus.WithFields(logrus.Fields{
						"node":   device.ID.Node,
						"device": device.Name,
						"state":  state,
					}).Infof("logged device values")
				}
			}
		}
		return err
	}
}

func appendNodeTypeOnTags(tags map[string]string, uuid string) {
	nodeType, ok := nodesList[uuid]
	if !ok {
		return
	}
	tags["node-type"] = nodeType
}

func worker(stop chan struct{}, queueChan chan func()) {
	logrus.Info("Starting worker")
	for {
		select {
		case <-stop:
			logrus.Info("stopping worker")
			return
		case fn := <-queueChan:
			fn()
		}
	}
}

func updatedConfig(stop chan struct{}, queueChan chan func()) func(data json.RawMessage) error {
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
		go worker(stop, queueChan)

		logrus.Infof("Config is now: %#v", config)
		return nil
	}
}

// InitClient makes a new influx db client.
func InitClient() (client.Client, error) {
	return client.NewHTTPClient(client.HTTPConfig{
		Addr:     fmt.Sprintf("http://%s:%s", config.Host, config.Port),
		Username: config.Username,
		Password: config.Password,
	})
}

func write(tags map[string]string, fields map[string]interface{}) error {
	for k, v := range fields {
		if v, ok := v.(bool); ok {
			if v {
				fields[k] = 1
			} else {
				fields[k] = 0
			}
		}
		if v, ok := v.(int); ok {
			fields[k] = float64(v)
		}
		if v == nil {
			delete(fields, k)
		}
	}

	if len(fields) == 0 {
		return fmt.Errorf("no loggable values")
	}

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
