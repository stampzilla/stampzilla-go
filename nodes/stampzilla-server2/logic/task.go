package logic

import (
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/models/devices"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/websocket"
)

type Task struct {
	Name_    string   `json:"name"`
	Uuid_    string   `json:"uuid"`
	Actions  []string `json:"actions"`
	cronId   int
	CronWhen string `json:"when"`
	sync.RWMutex
	SavedStateStore *SavedStateStore
	sender          websocket.Sender
}

func (t *Task) SetUuid(uuid string) {
	t.Lock()
	t.Uuid_ = uuid
	t.Unlock()
}
func (r *Task) Uuid() string {
	r.RLock()
	defer r.RUnlock()
	return r.Uuid_
}
func (r *Task) Name() string {
	r.RLock()
	defer r.RUnlock()
	return r.Name_
}
func (r *Task) CronId() int {
	r.RLock()
	defer r.RUnlock()
	return r.cronId
}

func (t *Task) Run() {
	t.RLock()
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
	t.RUnlock()
}

func (t *Task) AddAction(uuid string) {
	t.Lock()
	t.Actions = append(t.Actions, uuid)
	t.Unlock()
}
