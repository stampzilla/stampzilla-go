package handlers

import (
	"encoding/json"
	"strings"

	"github.com/Sirupsen/logrus"
	log "github.com/cihub/seelog"
	serverprotocol "github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/protocol"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/websocket"
	"github.com/stampzilla/stampzilla-go/protocol"
)

// Devices contains the websocket handler for devices
type Devices struct {
	Devices *serverprotocol.Devices `inject:""`
	Nodes   *serverprotocol.Nodes   `inject:""`
	Router  *websocket.Router       `inject:""`
	Clients *websocket.Clients      `inject:""`
}

// Start starts the websocket handler for devices by register its routes and client connect handler
func (d *Devices) Start() {

	//wh.Router.AddRoute("cmd", wh.RunCommand)

	d.Router.AddClientConnectHandler(func() *websocket.Message {
		return &websocket.Message{Type: "devices/all", Data: d.jsonRawMessage(d.Devices.All())}
	})

	d.Router.AddRoute("devices/set", d.set)
	d.Router.AddRoute("device/config/set", d.setConfig)

}
func (d *Devices) sendAllDevices() {
	go d.Clients.SendToAll("devices/all", d.Devices.All())
}

// SendSingleDevice sends an update to webgui clients with a single device update
func (d *Devices) SendSingleDevice(device interface{}) {
	go d.Clients.SendToAll("devices/single", device)
}

func (d *Devices) set(msg *websocket.Message) {
	type message struct {
		UUID string `json:"uuid"`
		Name string `json:"name"`
		Tags string `json:"tags"`
	}
	var data message

	err := json.Unmarshal(msg.Data, &data)
	if err != nil {
		log.Error("Failed to decode devices/set from websocket: ", err)
		return
	}

	device := d.Devices.ByUuid(data.UUID)
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

		go d.Clients.SendToAll("devices/single", device)

		d.Devices.SaveToFile("devices.json")
	}
}

func (d *Devices) setConfig(msg *websocket.Message) {
	type message struct {
		Device    string      `json:"device"`
		Parameter string      `json:"parameter"`
		Value     interface{} `json:"value"`
	}
	var data message

	err := json.Unmarshal(msg.Data, &data)
	if err != nil {
		log.Error("Failed to decode device/config/set from websocket: ", err)
		return
	}

	device := d.Devices.ByUuid(data.Device)
	if device == nil {
		logrus.Errorf("Received config but device (%s) was not found", data.Device)
		return
	}

	devid := strings.SplitN(data.Device, ".", 2)
	if len(devid) < 2 {
		logrus.Errorf("Received config but could not split device id (%s) ", data.Device)
		return
	}

	node := d.Nodes.ByUuid(devid[0])
	if node == nil {
		logrus.Errorf("Received config but node (%s) was not found", devid[0])
		return
	}

	cfg := protocol.DeviceConfigSet{
		Device: devid[1],
		ID:     data.Parameter,
		Value:  data.Value,
	}

	u := protocol.NewUpdateWithData(protocol.TypeDeviceConfigSet, cfg)
	err = serverprotocol.WriteUpdate(node, u)
	if err != nil {
		logrus.Errorf("Received config but failed to save: %s", err.Error())
		return
	}

	return
}

func (d *Devices) jsonRawMessage(data interface{}) json.RawMessage {
	msg, err := json.Marshal(data)
	if err != nil {
		log.Error(err)
		return nil
	}
	return msg
}
