package main

import (
	"encoding/json"

	log "github.com/cihub/seelog"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/logic"
	serverprotocol "github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/protocol"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/websocket"
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
		return &websocket.Message{Type: "all", Data: wh.jsonRawMessage(wh.Nodes.All())}
	})

}

func (wh *WebsocketHandler) jsonRawMessage(data interface{}) json.RawMessage {
	msg, err := json.Marshal(data)
	if err != nil {
		log.Error(err)
		return nil
	}
	return msg
}
func (wh *WebsocketHandler) SchedulerRemoveTask(msg *websocket.Message) {
	type Task struct {
		Uuid string `json:"uuid"`
	}
	task := &Task{}
	err := json.Unmarshal(msg.Data, task)
	if err != nil {
		log.Error(err)
		return
	}

	log.Info("Removing scheduled task ", task.Uuid)
	if err := wh.Scheduler.RemoveTask(task.Uuid); err != nil {
		log.Error(err)
		return
	}
	wh.Clients.SendToAll("scheduleall", wh.Scheduler.Tasks())
}

func (wh *WebsocketHandler) SchedulerEditTask(msg *websocket.Message) {
	wh.Clients.SendToAll("scheduleall", wh.Scheduler.Tasks())
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
func (wh *WebsocketHandler) RunCommand(msg *websocket.Message) {
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

func (wh *WebsocketHandler) SendAllNodes() {
	go wh.Clients.SendToAll("all", wh.Nodes.All())
}
func (wh *WebsocketHandler) SendSingleNode(uuid string) {
	go wh.Clients.SendToAll("singlenode", wh.Nodes.ByUuid(uuid))
}
