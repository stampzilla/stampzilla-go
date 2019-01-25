package logic

import (
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models/devices"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/websocket"
)

// Task is a task that can be scheduled using scheduler
type Task struct {
	XName   string   `json:"name"`
	XUuid   string   `json:"uuid"`
	Actions []string `json:"actions"`
	cronID  int64
	When    string `json:"when"`
	sync.RWMutex
	SavedStateStore *SavedStateStore
	sender          websocket.Sender
}

// SetUuid sets ths uuid on the task
func (t *Task) SetUuid(uuid string) {
	t.Lock()
	t.XUuid = uuid
	t.Unlock()
}

// SetWhen sets when the task should be run in cron syntax
func (t *Task) SetWhen(when string) {
	t.Lock()
	t.When = when
	t.Unlock()
}

// Uuid returns tasks uuid
func (r *Task) Uuid() string {
	r.RLock()
	defer r.RUnlock()
	return r.XUuid
}
func (r *Task) Name() string {
	r.RLock()
	defer r.RUnlock()
	return r.XName
}
func (r *Task) CronId() int64 {
	r.RLock()
	defer r.RUnlock()
	return r.cronID
}

func (t *Task) Run() {
	t.RLock()
	defer t.RUnlock()
	for _, id := range t.Actions {
		stateList := t.SavedStateStore.Get(id)
		if stateList == nil {
			logrus.Errorf("SavedState %s does not exist", id)
			return
		}
		devicesByNode := make(map[string]map[devices.ID]devices.State)
		for id, state := range stateList.State {
			if devicesByNode[id.Node] == nil {
				devicesByNode[id.Node] = make(map[devices.ID]devices.State)
			}
			devicesByNode[id.Node][id] = state
		}
		for nodeID, devs := range devicesByNode {
			logrus.WithFields(logrus.Fields{
				"to": nodeID,
			}).Debug("Send state change request to node")
			err := t.sender.SendToID(nodeID, "state-change", devs)
			if err != nil {
				logrus.Error("logic: error sending state-change to node: ", err)
				continue
			}
		}
	}
}

func (t *Task) AddAction(uuid string) {
	t.Lock()
	t.Actions = append(t.Actions, uuid)
	t.Unlock()
}
