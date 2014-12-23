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

func (wh *WebsocketHandler) jsonDecode(str string) *websocket.Message {
	var msg *websocket.Message
	err := json.Unmarshal([]byte(str), &msg)
	if err != nil {
		log.Error(err)
		return nil
	}
	return msg
}
func (wh *WebsocketHandler) SchedulerRemoveTask(str string) {
	msg := wh.jsonDecode(str)
	if str, ok := msg.Data.(string); ok {
		log.Info("Removing scheduled task ", str)
		if err := wh.Scheduler.RemoveTask(str); err != nil {
			log.Error(err)
			return
		}
		wh.Clients.SendToAll(&websocket.Message{Type: "scheduleall", Data: wh.Scheduler.Tasks()})
	}
}

func (wh *WebsocketHandler) SchedulerEditTask(msg *websocket.Message) {

	wh.Clients.SendToAll(&websocket.Message{Type: "scheduleall", Data: wh.Scheduler.Tasks()})
}
func (wh *WebsocketHandler) SchedulerAddTask(msg string) {

	//TODO do the json decore here now when msg is a string!
	//name := ""
	//uuid := ""
	//if task, ok := msg.Data.(map[string]interface{}); ok {
	//if name, ok = task["name"].(string); !ok {
	//return
	//}
	//if uuid, ok = task["uuid"].(string); !ok {
	//return
	//}

	//t := wh.Scheduler.AddTask(name)

	////Set the uuid from json if it exists. Otherwise use the generated one
	//if uuid != "" {
	//t.SetUuid(uuid)
	//}
	////for _, cond := range task.Actions {
	////t.AddAction(cond)
	////}
	//////Schedule the task!
	////t.Schedule(task.CronWhen)

	//}

	//wh.Clients.SendToAll(&websocket.Message{Type: "scheduleall", Data: wh.Scheduler.Tasks()})
}
func (wh *WebsocketHandler) RunCommand(str string) {
	msg := wh.jsonDecode(str)
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
