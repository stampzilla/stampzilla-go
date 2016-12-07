package handlers

import (
	"encoding/json"

	log "github.com/cihub/seelog"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/protocol"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/websocket"
)

type Devices struct {
	Devices *protocol.Devices  `inject:""`
	Router  *websocket.Router  `inject:""`
	Clients *websocket.Clients `inject:""`
}

func (wsr *Devices) Start() {

	//wh.Router.AddRoute("cmd", wh.RunCommand)

	wsr.Router.AddClientConnectHandler(func() *websocket.Message {
		return &websocket.Message{Type: "devices/all", Data: wsr.jsonRawMessage(wsr.Devices.All())}
	})

}
func (wh *Devices) SendAllDevices() {
	go wh.Clients.SendToAll("devices/all", wh.Devices.All())
}
func (wh *Devices) SendSingleDevice(device interface{}) {
	go wh.Clients.SendToAll("devices/single", device)
}

func (wsr *Devices) jsonRawMessage(data interface{}) json.RawMessage {
	msg, err := json.Marshal(data)
	if err != nil {
		log.Error(err)
		return nil
	}
	return msg
}
