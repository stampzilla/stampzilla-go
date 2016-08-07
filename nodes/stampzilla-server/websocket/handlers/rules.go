package handlers

import (
	"encoding/json"

	log "github.com/cihub/seelog"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/logic"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/websocket"
)

type Rules struct {
	Logic   *logic.Logic       `inject:""`
	Router  *websocket.Router  `inject:""`
	Clients *websocket.Clients `inject:""`
}

func (wsr *Rules) Start() {

	//wh.Router.AddRoute("cmd", wh.RunCommand)

	wsr.Router.AddClientConnectHandler(func() *websocket.Message {
		return &websocket.Message{Type: "rules/all", Data: wsr.jsonRawMessage(wsr.Logic.Rules())}
	})

}

func (wsr *Rules) jsonRawMessage(data interface{}) json.RawMessage {
	msg, err := json.Marshal(data)
	if err != nil {
		log.Error(err)
		return nil
	}
	return msg
}

//func (wsa *WebsocketActions) RunCommand(msg *websocket.Message) {
////msg := wh.jsonDecode(str)
//node := wh.Nodes.Search(msg.To)
//if node != nil {
//jsonToSend, err := json.Marshal(&msg.Data)
//if err != nil {
//log.Error(err)
//return
//}
//node.Write(jsonToSend)
//}
//}
