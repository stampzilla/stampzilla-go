package handlers

import (
	"encoding/json"

	log "github.com/cihub/seelog"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/logic"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/websocket"
)

type Actions struct {
	Actions *logic.ActionService `inject:""`
	Router  *websocket.Router    `inject:""`
	Clients *websocket.Clients   `inject:""`
}

func (wsa *Actions) Start() {

	//wh.Router.AddRoute("cmd", wh.RunCommand)

	wsa.Router.AddClientConnectHandler(func() *websocket.Message {
		return &websocket.Message{Type: "actions/all", Data: wsa.jsonRawMessage(wsa.Actions.Get())}
	})

}

func (wsa *Actions) jsonRawMessage(data interface{}) json.RawMessage {
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
