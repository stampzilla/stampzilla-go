package logic

import (
	"encoding/json"
	"sync/atomic"
	"testing"
	"time"

	"github.com/olahol/melody"
	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models/devices"
	"github.com/stretchr/testify/assert"
)

type mockSender struct {
	Devices *devices.List
	count   int64
}

func NewMockSender() *mockSender {
	return &mockSender{
		Devices: devices.NewList(),
	}
}

func (mss *mockSender) Count() int64 {
	return atomic.LoadInt64(&mss.count)
}

func (mss *mockSender) SendToID(to string, msgType string, data interface{}) error {
	atomic.AddInt64(&mss.count, 1)
	for k, v := range data.(map[devices.ID]devices.State) {
		mss.Devices.Add(&devices.Device{
			ID:    k,
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
	savedState := NewSavedStateStore()
	l := New(savedState, syncer)
	l.Load()
	//spew.Dump(l.Rules)
	jsonData, err := json.MarshalIndent(l.Rules, "", "\t")
	assert.NoError(t, err)
	t.Log(string(jsonData))
}

func TestEvaluateRules(t *testing.T) {

	syncer := NewMockSender()

	savedState := NewSavedStateStore()
	l := New(savedState, syncer)
	r := l.AddRule("test")
	r.Expression_ = `devices["node.id"].on == true`

	l.updateDevice(&devices.Device{
		ID: devices.ID{
			Node: "node",
			ID:   "id",
		},
		State: devices.State{
			"on": true,
		},
	})

	l.EvaluateRules()

	assert.Equal(t, true, l.Rules[r.Uuid()].Active())

	l.updateDevice(&devices.Device{
		ID: devices.ID{
			Node: "node",
			ID:   "id",
		},
		State: devices.State{
			"on": false,
		},
	})
	l.EvaluateRules()
	assert.Equal(t, false, l.Rules[r.Uuid()].Active())
}

func TestEvaluateRulesCanceledIfNotActive(t *testing.T) {

	logrus.SetLevel(logrus.DebugLevel)
	syncer := NewMockSender()

	savedState := NewSavedStateStore()
	savedState.State["uuid"] = &SavedState{
		Name: "testname",
		UUID: "uuid",
		State: map[devices.ID]devices.State{
			devices.ID{Node: "node", ID: "id"}: devices.State{
				"on": false,
				"a":  true,
				"b":  true,
			},
		},
	}
	l := New(savedState, syncer)
	r := l.AddRule("test")
	r.Expression_ = `devices["node.id"].on == true`
	r.Actions_ = []string{
		"uuid",
		"20ms",
		"uuid",
	}

	l.updateDevice(&devices.Device{
		ID: devices.ID{
			Node: "node",
			ID:   "id",
		},
		State: devices.State{
			"on": true,
		},
	})

	l.EvaluateRules()
	assert.Equal(t, true, l.Rules[r.Uuid()].Active())

	time.Sleep(10 * time.Millisecond)

	l.updateDevice(&devices.Device{
		ID: devices.ID{
			Node: "node",
			ID:   "id",
		},
		State: devices.State{
			"on": false,
		},
	})
	l.EvaluateRules()
	assert.Equal(t, false, l.Rules[r.Uuid()].Active())

	l.Wait()

	assert.Equal(t, int64(1), syncer.Count())
}
