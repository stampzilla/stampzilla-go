package handlers

import (
	"encoding/json"
	"strings"

	"github.com/Sirupsen/logrus"
	log "github.com/cihub/seelog"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/protocol"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/websocket"
)

type Devices struct {
	Devices *protocol.Devices  `inject:""`
	Nodes   *protocol.Nodes    `inject:""`
	Router  *websocket.Router  `inject:""`
	Clients *websocket.Clients `inject:""`
}

func (wsr *Devices) Start() {

	//wh.Router.AddRoute("cmd", wh.RunCommand)

	wsr.Router.AddClientConnectHandler(func() *websocket.Message {
		return &websocket.Message{Type: "devices/all", Data: wsr.jsonRawMessage(wsr.Devices.All())}
	})

	wsr.Router.AddRoute("devices/set", wsr.Set)
	wsr.Router.AddRoute("device/config/set", wsr.SetConfig)

}
func (wh *Devices) SendAllDevices() {
	go wh.Clients.SendToAll("devices/all", wh.Devices.All())
}
func (wh *Devices) SendSingleDevice(device interface{}) {
	go wh.Clients.SendToAll("devices/single", device)
}

func (wh *Devices) Set(msg *websocket.Message) {
	type message struct {
		Uuid string `json:"uuid"`
		Name string `json:"name"`
		Tags string `json:"tags"`
	}
	var data message

	json.Unmarshal(msg.Data, &data)

	device := wh.Devices.ByUuid(data.Uuid)
	if device != nil {
		if data.Name != "" {
			device.Name = data.Name
		}

		if data.Tags != "" {
			tags := strings.Split(data.Tags, ",")
			for k, v := range tags {
				tags[k] = strings.TrimSpace(v)
			}
			device.Tags = tags
		}

		go wh.Clients.SendToAll("devices/single", device)

		wh.Devices.SaveToFile("devices.json")
	}
}

func (wh *Devices) SetConfig(msg *websocket.Message) {
	type message struct {
		Device    string      `json:"device"`
		Parameter string      `json:"parameter"`
		Value     interface{} `json:"value"`
	}
	var data message

	json.Unmarshal(msg.Data, &data)

	device := wh.Devices.ByUuid(data.Device)
	if device == nil {
		logrus.Errorf("Received config but device (%s) was not found", data.Device)
	}
	//device.SetConfig(data.Parameter, data.Value)

	devid := strings.SplitN(data.Device, ".", 2)
	if len(devid) < 2 {
		logrus.Errorf("Received config but could not split device id (%s) ", data.Device)
	}

	node := wh.Nodes.ByUuid(devid[0])
	if node == nil {
		logrus.Errorf("Received config but node (%s) was not found", devid[0])
	}

	err := node.SaveConfig(devid[1], data.Parameter, data.Value)
	if err != nil {
		logrus.Errorf("Received config but failed to save: %s", err.Error())
	}
}

func (wsr *Devices) jsonRawMessage(data interface{}) json.RawMessage {
	msg, err := json.Marshal(data)
	if err != nil {
		log.Error(err)
		return nil
	}
	return msg
}
