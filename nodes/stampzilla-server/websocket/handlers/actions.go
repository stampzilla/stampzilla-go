package handlers

import (
	"encoding/json"

	log "github.com/cihub/seelog"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/logic"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/websocket"
)

type Actions struct {
	Actions *logic.ActionService `inject:""`
	Logic   *logic.Logic         `inject:""`
	Router  *websocket.Router    `inject:""`
	Clients *websocket.Clients   `inject:""`
}

func (wsa *Actions) Start() {

	wsa.Router.AddRoute("actions/run", wsa.Run)

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

func (wsa *Actions) Run(msg *websocket.Message) {
	var uuid string
	json.Unmarshal(msg.Data, &uuid)

	a := wsa.Actions.GetByUuid(uuid)

	if a == nil {
		log.Errorf("Action \"%s\" was not found", uuid)
		return
	}

	a.Run(wsa.Logic.ActionProgressChan)

	//msg := wh.jsonDecode(str)
	//node := wh.Nodes.Search(msg.To)
	//if node != nil {
	//jsonToSend, err := json.Marshal(&msg.Data)
	//if err != nil {
	//log.Error(err)
	//return
	//}
	//node.Write(jsonToSend)
	//}
}
