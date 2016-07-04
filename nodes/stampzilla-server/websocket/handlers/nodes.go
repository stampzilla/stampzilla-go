package handlers

import (
	"encoding/json"

	log "github.com/cihub/seelog"
	serverprotocol "github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/protocol"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/websocket"
)

type Nodes struct {
	Nodes   *serverprotocol.Nodes `inject:""`
	Router  *websocket.Router     `inject:""`
	Clients *websocket.Clients    `inject:""`
}

func (wh *Nodes) Start() {

	// cmd
	wh.Router.AddRoute("nodes/cmd", wh.RunCommand)

	wh.Router.AddClientConnectHandler(func() *websocket.Message {
		return &websocket.Message{Type: "nodes/all", Data: wh.jsonRawMessage(wh.Nodes.All())}
	})

}

func (wh *Nodes) jsonRawMessage(data interface{}) json.RawMessage {
	msg, err := json.Marshal(data)
	if err != nil {
		log.Error(err)
		return nil
	}
	return msg
}

func (wh *Nodes) RunCommand(msg *websocket.Message) {
	//msg := wh.jsonDecode(str)
	node := wh.Nodes.Search(msg.To)
	if node != nil {
		jsonToSend, err := json.Marshal(&msg.Data)
		if err != nil {
			log.Error(err)
			return
		}
		node.Write(jsonToSend)
	}
}

func (wh *Nodes) SendAllNodes() {
	go wh.Clients.SendToAll("nodes/all", wh.Nodes.All())
}
func (wh *Nodes) SendSingleNode(uuid string) {
	go wh.Clients.SendToAll("nodes/single", wh.Nodes.ByUuid(uuid))
}
func (wh *Nodes) SendDisconnectedNode(uuid string) {
	go wh.Clients.SendToAll("nodes/disconnected", uuid)
}
