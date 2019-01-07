package logic

import (
	"encoding/json"
	"testing"

	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/models/devices"
	"github.com/stretchr/testify/assert"
)

type mockStateSyncer struct {
	Devices *devices.List
}

func NewStateSyncer() *mockStateSyncer {
	return &mockStateSyncer{
		Devices: devices.NewList(),
	}
}

func (mss mockStateSyncer) SyncState(list map[string]devices.State) {
	for id, state := range list {
		dev := mss.Devices.GetUnique(id)
		if dev == nil {
			return
		}
		dev.SyncState(state)
	}
}

func TestLoadRulesFromFile(t *testing.T) {

	syncer := NewStateSyncer()
	l := NewLogic(syncer)
	l.Load("rules.json")
	//spew.Dump(l.Rules)
	jsonData, err := json.MarshalIndent(l.Rules, "", "\t")
	assert.NoError(t, err)
	t.Log(string(jsonData))
}

func TestEvaluateRules(t *testing.T) {

	syncer := NewStateSyncer()

	l := NewLogic(syncer)
	r := l.AddRule("test")
	r.Expression_ = `devices["node.id"].on == true`

	l.UpdateDevice(&devices.Device{
		Node: "node",
		ID:   "id",
		State: devices.State{
			"on": true,
		},
	})

	l.EvaluateRules()

	assert.Equal(t, true, l.Rules[r.Uuid()].Active())

	l.UpdateDevice(&devices.Device{
		Node: "node",
		ID:   "id",
		State: devices.State{
			"on": false,
		},
	})
	l.EvaluateRules()
	assert.Equal(t, false, l.Rules[r.Uuid()].Active())
}
