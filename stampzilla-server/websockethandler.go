package main

import (
	"encoding/json"

	log "github.com/cihub/seelog"
	serverprotocol "github.com/stampzilla/stampzilla-go/stampzilla-server/protocol"
	"github.com/stampzilla/stampzilla-go/stampzilla-server/websocket"
)

type WebsocketHandler struct {
	//Logic     *logic.Logic          `inject:""`
	//Scheduler *logic.Scheduler      `inject:""`
	Nodes  *serverprotocol.Nodes `inject:""`
	Router *websocket.Router     `inject:""`
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
func (wh *WebsocketHandler) Start() {

	// cmd
	wh.Router.AddRoute("cmd", wh.RunCommand)

}
