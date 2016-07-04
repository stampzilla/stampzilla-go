package handlers

import (
	"encoding/json"

	log "github.com/cihub/seelog"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/logic"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/websocket"
)

type Schedule struct {
	Scheduler *logic.Scheduler   `inject:""`
	Router    *websocket.Router  `inject:""`
	Clients   *websocket.Clients `inject:""`
}

func (wh *Schedule) Start() {

	// cmd
	//wh.Router.AddRoute("cmd", wh.RunCommand)

	wh.Router.AddClientConnectHandler(func() *websocket.Message {
		return &websocket.Message{Type: "schedule/all", Data: wh.jsonRawMessage(wh.Scheduler.Tasks())}
	})

}

func (wh *Schedule) jsonRawMessage(data interface{}) json.RawMessage {
	msg, err := json.Marshal(data)
	if err != nil {
		log.Error(err)
		return nil
	}
	return msg
}
func (wh *Schedule) SchedulerRemoveTask(msg *websocket.Message) {
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
	wh.Clients.SendToAll("schedule/all", wh.Scheduler.Tasks())
}

func (wh *Schedule) SchedulerEditTask(msg *websocket.Message) {
	wh.Clients.SendToAll("schedule/all", wh.Scheduler.Tasks())
}
func (wh *Schedule) SchedulerAddTask(msg string) {

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

	//wh.Clients.SendToAll(&websocket.Message{Type: "schedule/all", Data: wh.Scheduler.Tasks()})
}
