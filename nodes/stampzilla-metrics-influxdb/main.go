package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/influxdb/influxdb/client/v2"
	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/pkg/node"
	"github.com/stampzilla/stampzilla-go/pkg/websocket"
)

type Config struct {
	Host     string
	Port     string
	Username string
	Password string
	Database string
}

var config = &Config{}

var influxClient client.Client

func main() {

	client := websocket.New()
	node := node.New(client)

	node.OnConfig(updatedConfig)
	node.On("devices", onDevices)
	err := node.Connect()

	if err != nil {
		logrus.Error(err)
		return
	}
	node.Subscribe("devices")

	influxClient, err = InitClient()
	if err != nil {
		logrus.Error(err)
		return
	}

	defer influxClient.Close()
	node.Wait()
}
func onDevices(data json.RawMessage) error {
	logrus.Info("devices incoming:", string(data))
	return nil
}

func updatedConfig(data json.RawMessage) error {
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

	// stop worker

	influxClient, err = InitClient()
	if err != nil {
		return err
	}

	// start worker

	logrus.Infof("Config is now: %#v", config)
	return nil
}

func InitClient() (client.Client, error) {
	return client.NewHTTPClient(client.HTTPConfig{
		Addr:     fmt.Sprintf("http://%s:%s", config.Host, config.Port),
		Username: config.Username,
		Password: config.Password,
	})
}

func write() {
	// Create a new point batch
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  config.Database,
		Precision: "s",
	})
	if err != nil {
		log.Fatal(err)
	}

	// Create a point and add to batch
	tags := map[string]string{"cpu": "cpu-total"}
	fields := map[string]interface{}{
		"idle":   10.1,
		"system": 53.3,
		"user":   46.6,
	}

	pt, err := client.NewPoint("cpu_usage", tags, fields, time.Now())
	if err != nil {
		log.Fatal(err)
	}
	bp.AddPoint(pt)

	// Write the batch
	if err := influxClient.Write(bp); err != nil {
		log.Fatal(err)
	}

}
