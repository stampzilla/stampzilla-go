package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-google-assistant/googleassistant"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models/devices"
	"github.com/stampzilla/stampzilla-go/pkg/node"
)

func main() {
	config := &Config{}

	node := node.New("google-assistant")

	deviceList := devices.NewList()
	smartHomeHandler := NewSmartHomeHandler(node, deviceList)

	node.OnConfig(updatedConfig(config))
	wait := node.WaitForFirstConfig()

	err := node.Connect()
	if err != nil {
		logrus.Error(err)
		return
	}

	wait()
	node.On("devices", onDevices(node, config, deviceList))

	handleIntent := func(r *googleassistant.Request) (interface{}, error) {
		switch r.Inputs.Intent() {
		case googleassistant.SyncIntent:
			return smartHomeHandler.syncHandler(node.UUID, r), nil
		case googleassistant.ExecuteIntent:
			return smartHomeHandler.executeHandler(r), nil
		case googleassistant.QueryIntent:
			return smartHomeHandler.queryHandler(r), nil
		}
		return nil, fmt.Errorf("Unknown intent")
	}

	node.OnCloudRequest(func(req *http.Request) (*http.Response, error) {

		dec := json.NewDecoder(req.Body)
		defer req.Body.Close()
		r := &googleassistant.Request{}

		err = dec.Decode(r)
		if err != nil {
			return nil, err
		}

		logrus.Info("Intent: ", r.Inputs.Intent())
		logrus.Debug("Request:", spew.Sdump(r))

		data, err := handleIntent(r)
		if err != nil {
			return nil, err
		}

		jsonBytes, err := json.MarshalIndent(data, "", "    ")
		if err != nil {
			return nil, err
		}

		return &http.Response{
			StatusCode: 200,
			ProtoMajor: 1,
			ProtoMinor: 0,
			Request:    req,
			Header: http.Header{
				"Content-Type": []string{"application/json"},
			},
			Body:          ioutil.NopCloser(bytes.NewReader(jsonBytes)),
			ContentLength: -1,
		}, nil
	})

	go func() {
		time.Sleep(5 * time.Second)
		logrus.Info("Syncing devices to google")
		requestSync(node, config.APIKey)
	}()

	node.Wait()
}

func onDevices(node *node.Node, config *Config, deviceList *devices.List) func(data json.RawMessage) error {
	return func(data json.RawMessage) error {
		list := devices.NewList()
		err := json.Unmarshal(data, list)
		if err != nil {
			return err
		}

		changes := 0
		for _, dev := range list.All() {
			// Only list devices that has the cloud label
			if _, ok := dev.Label("cloud"); !ok {
				continue
			}

			old := deviceList.Get(dev.ID)
			if old == nil {
				deviceList.Add(dev)
				changes++
				continue
			}

			if old.Name != dev.Name {
				old.Name = dev.Name
				changes++
			}
			if old.Alias != dev.Alias {
				old.Alias = dev.Alias
				changes++
			}
		}

		toRemove := []devices.ID{}
		for _, v := range deviceList.All() {
			dev := list.Get(v.ID)
			if dev == nil {
				toRemove = append(toRemove, v.ID)
				continue
			}
			if _, ok := dev.Label("cloud"); !ok {
				toRemove = append(toRemove, v.ID)
			}
		}
		for _, id := range toRemove {
			changes++
			deviceList.Remove(id)
		}

		if changes > 0 {
			logrus.Infof("Device list has changed, it now contains %d devices", deviceList.Len())
			requestSync(node, config.APIKey)
		}

		return nil
	}
}
func updatedConfig(config *Config) func(data json.RawMessage) error {
	return func(data json.RawMessage) error {
		logrus.Debug("Received config from server:", string(data))
		return json.Unmarshal(data, &config)
	}
}

func requestSync(node *node.Node, apiKey string) {
	u := fmt.Sprintf("https://homegraph.googleapis.com/v1/devices:requestSync")

	body := bytes.NewBufferString("{agent_user_id: \"" + node.UUID + "\"}")
	req, err := http.NewRequest("POST", u, body)
	if err != nil {
		logrus.Error("requestsync: ", err)
		return
	}

	req.Header.Add("Content-Type", "application/json")

	resp, err := node.SendThruCloud("google-assistant", req)
	if err != nil {
		logrus.Error("requestsync: ", err)
		return
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.Error("requestsync: ", err)
		return
	}

	logrus.Debug("requestSync response:", string(data))
}
