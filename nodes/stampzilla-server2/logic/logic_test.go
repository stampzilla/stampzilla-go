package logic

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/olahol/melody"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/models/devices"
	"github.com/stretchr/testify/assert"
)

type mockSender struct {
	Devices *devices.List
}

func NewMockSender() *mockSender {
	return &mockSender{
		Devices: devices.NewList(),
	}
}

func (mss *mockSender) SendToID(to string, msgType string, data interface{}) error {
	for k, v := range data.(map[string]devices.State) {
		id := strings.Split(k, ".")
		mss.Devices.Add(&devices.Device{
			Node:  to,
			ID:    id[1],
			State: v,
		})
	}
	return nil
}
func (mss *mockSender) SendToProtocol(to string, msgType string, data interface{}) error {

	return nil
}
func (mss *mockSender) BroadcastWithFilter(msgType string, data interface{}, fn func(*melody.Session) bool) error {
	return nil
}

func TestLoadRulesFromFile(t *testing.T) {

	syncer := NewMockSender()
	l := New(syncer)
	l.Load("rules.json")
	//spew.Dump(l.Rules)
	jsonData, err := json.MarshalIndent(l.Rules, "", "\t")
	assert.NoError(t, err)
	t.Log(string(jsonData))
}

func TestEvaluateRules(t *testing.T) {

	syncer := NewMockSender()

	l := New(syncer)
	r := l.AddRule("test")
	r.Expression_ = `devices["node.id"].on == true`

	l.updateDevice(&devices.Device{
		Node: "node",
		ID:   "id",
		State: devices.State{
			"on": true,
		},
	})

	l.EvaluateRules()

	assert.Equal(t, true, l.Rules[r.Uuid()].Active())

	l.updateDevice(&devices.Device{
		Node: "node",
		ID:   "id",
		State: devices.State{
			"on": false,
		},
	})
	l.EvaluateRules()
	assert.Equal(t, false, l.Rules[r.Uuid()].Active())
}
