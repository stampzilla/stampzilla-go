package logic

import (
	"sync"

	"github.com/google/cel-go/cel"
	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/models/devices"
	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/websocket"
)

// Task is a task that can be scheduled using scheduler.
type Task struct {
	XName       string   `json:"name"`
	XUuid       string   `json:"uuid"`
	Actions     []string `json:"actions"`
	cronID      int64
	When        string `json:"when"`
	Enabled     bool   `json:"enabled"`
	Expression_ string `json:"expression"`
	logic       *Logic
	ast         *cel.Ast
	sync.RWMutex
	savedStateStore *SavedStateStore
	sender          websocket.Sender
}

// SetUuid sets ths uuid on the task.
func (t *Task) SetUuid(uuid string) {
	t.Lock()
	t.XUuid = uuid
	t.Unlock()
}

// SetWhen sets when the task should be run in cron syntax.
func (t *Task) SetWhen(when string) {
	t.Lock()
	t.When = when
	t.Unlock()
}
func (t *Task) Expression() string {
	t.RLock()
	defer t.RUnlock()
	return t.Expression_
}

// Uuid returns tasks uuid.
func (t *Task) Uuid() string {
	t.RLock()
	defer t.RUnlock()
	return t.XUuid
}

func (t *Task) Name() string {
	t.RLock()
	defer t.RUnlock()
	return t.XName
}

func (t *Task) CronId() int64 {
	t.RLock()
	defer t.RUnlock()
	return t.cronID
}

func (t *Task) Run() {
	t.RLock()
	defer t.RUnlock()
	if !t.Enabled {
		logrus.Debugf("logic: scheduledtask %s (%s) not enabled. skipping", t.Name(), t.Uuid())
		return
	}

	if exp := t.Expression(); exp != "" {
		rules := make(map[string]bool)
		for _, v := range t.logic.Rules {
			rules[v.Uuid()] = v.Active()
		}
		b, err := eval(exp, t.logic.devices, nil, t.ast)
		if err != nil {
			logrus.Error(err)
			return
		}
		if !b {
			return
		}
	}

	for _, id := range t.Actions {
		stateList := t.savedStateStore.Get(id)
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
