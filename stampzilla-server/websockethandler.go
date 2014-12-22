package main

import (
	"encoding/json"

	log "github.com/cihub/seelog"
	"github.com/stampzilla/stampzilla-go/stampzilla-server/logic"
	serverprotocol "github.com/stampzilla/stampzilla-go/stampzilla-server/protocol"
	"github.com/stampzilla/stampzilla-go/stampzilla-server/websocket"
)

type WebsocketHandler struct {
	//Logic     *logic.Logic          `inject:""`
	Scheduler *logic.Scheduler      `inject:""`
	Nodes     *serverprotocol.Nodes `inject:""`
	Router    *websocket.Router     `inject:""`
	Clients   *websocket.Clients    `inject:""`
}

func (wh *WebsocketHandler) Start() {

	// cmd
	wh.Router.AddRoute("cmd", wh.RunCommand)

	wh.Router.AddClientConnectHandler(func() *websocket.Message {
		return &websocket.Message{Type: "all", Data: wh.Nodes.All()}
	})

}

func (wh *WebsocketHandler) SchedulerRemoveTask(msg *websocket.Message) {
	if str, ok := msg.Data.(string); ok {
		if err := wh.Scheduler.RemoveTask(str); err != nil {
			log.Error(err)
			return
		}
		wh.Clients.SendToAll(&websocket.Message{Type: "scheduleall", Data: wh.Scheduler.Tasks()})
	}
}
func (wh *WebsocketHandler) RunCommand(msg *websocket.Message) {
	node := wh.Nodes.Search(msg.To)
	if node != nil {
		jsonToSend, err := json.Marshal(&msg.Data)
		if err != nil {
			log.Error(err)
			return
		}
		node.Conn().Write(jsonToSend)
	}
}

func (wh *WebsocketHandler) SendAllNodes() {
	wh.Clients.SendToAll(&websocket.Message{Type: "all", Data: wh.Nodes.All()})
}
func (wh *WebsocketHandler) SendSingleNode(uuid string) {
	wh.Clients.SendToAll(&websocket.Message{Type: "singlenode", Data: wh.Nodes.ByUuid(uuid)})
}
