package logic

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"testing"
	"time"

	"github.com/olahol/melody"
	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/models/devices"
	stypes "github.com/stampzilla/stampzilla-go/v2/pkg/types"
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

	assert.Equal(t, stypes.Duration(time.Minute*5), l.Rules["e8092b86-1261-44cd-ab64-38121df58a79"].For_)
	// spew.Dump(l.Rules)
	jsonData, err := json.MarshalIndent(l.Rules, "", "\t")
	assert.NoError(t, err)
	t.Log(string(jsonData))
}

func TestGetSetRules(t *testing.T) {
	syncer := NewMockSender()
	savedState := NewSavedStateStore()
	l := New(savedState, syncer)

	rules := Rules{
		"rule1": &Rule{},
		"rule2": &Rule{},
	}
	cnt := 0
	go func() {
		<-l.c
		cnt++
	}()
	l.SetRules(rules)

	assert.Equal(t, l.GetRules(), rules)
	assert.Equal(t, cnt, 1)
}

func TestEvaluateRules(t *testing.T) {
	syncer := NewMockSender()
	savedState := NewSavedStateStore()
	l := New(savedState, syncer)

	r := l.AddRule("test")
	r.Expression_ = `devices["node.id"].on == true`
	r.Enabled = true
	l.updateDevice(&devices.Device{
		ID: devices.ID{
			Node: "node",
			ID:   "id",
		},
		State: devices.State{
			"on": true,
		},
	})

	l.EvaluateRules(context.Background())

	assert.Equal(t, true, l.Rules[r.Uuid()].Active())

	ctx, cancel := context.WithCancel(context.Background())
	l.Start(ctx)
	l.UpdateDevice(&devices.Device{
		ID: devices.ID{
			Node: "node",
			ID:   "id",
		},
		State: devices.State{
			"on": false,
		},
	})
	cancel()
	l.Wait()
	// l.EvaluateRules(context.Background())
	assert.Equal(t, false, l.Rules[r.Uuid()].Active())
}

func TestEvaluateBrokenRules(t *testing.T) {
	syncer := NewMockSender()
	savedState := NewSavedStateStore()
	l := New(savedState, syncer)

	r := l.AddRule("test")
	r.Expression_ = `devices["node.id"].on =a= true`
	r.Enabled = true
	l.updateDevice(&devices.Device{
		ID: devices.ID{
			Node: "node",
			ID:   "id",
		},
		State: devices.State{
			"on": true,
		},
	})

	gotError := false
	l.OnReportState(func(id string, state devices.State) {
		if state["error"].(string) != "" {
			gotError = true
		}
	})
	l.EvaluateRules(context.Background())

	assert.Equal(t, false, l.Rules[r.Uuid()].Active())
	assert.Equal(t, true, gotError)
}

func TestEvaluateRulesCanceledIfNotActive(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	syncer := NewMockSender()

	savedState := NewSavedStateStore()
	savedState.State["uuid"] = &SavedState{
		Name: "testname",
		UUID: "uuid",
		State: map[devices.ID]devices.State{
			{Node: "node", ID: "id"}: {
				"on": false,
				"a":  true,
				"b":  true,
			},
		},
	}
	l := New(savedState, syncer)
	r := l.AddRule("test")
	r.Expression_ = `devices["node.id"].on == true`
	r.Enabled = true
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

	l.EvaluateRules(context.Background())
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
	l.EvaluateRules(context.Background())
	assert.Equal(t, false, l.Rules[r.Uuid()].Active())

	l.Wait()

	assert.Equal(t, int64(1), syncer.Count())
}

func TestEvaluateRulesWithFor(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	syncer := NewMockSender()

	devID := devices.ID{Node: "node", ID: "id"}
	savedState := NewSavedStateStore()
	savedState.State["uuid"] = &SavedState{
		Name: "testname",
		UUID: "uuid",
		State: map[devices.ID]devices.State{
			devID: {
				"on": false,
				"a":  true,
				"b":  true,
			},
		},
	}
	l := New(savedState, syncer)
	r := l.AddRule("test")
	r.Expression_ = `devices["node.id"].on == true`
	r.Enabled = true
	r.For_ = stypes.Duration(time.Millisecond * 20)
	r.Actions_ = []string{
		"uuid",
	}

	l.updateDevice(&devices.Device{
		ID: devID,
		State: devices.State{
			"on": true,
		},
	})
	l.EvaluateRules(context.Background())

	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, false, l.Rules[r.Uuid()].Active())
	time.Sleep(20 * time.Millisecond)
	assert.Equal(t, true, l.Rules[r.Uuid()].Active())

	l.updateDevice(&devices.Device{
		ID: devID,
		State: devices.State{
			"on": false,
		},
	})
	l.EvaluateRules(context.Background())
	assert.Equal(t, false, l.Rules[r.Uuid()].Active())

	l.Wait()

	assert.Equal(t, int64(1), syncer.Count())
}

// TestEvaluateRulesWithForTimeout asserts that we do not run the rule if it goes inactive before the "for" timeout.
func TestEvaluateRulesWithForTimeout(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	syncer := NewMockSender()

	devID := devices.ID{Node: "node", ID: "id"}
	savedState := NewSavedStateStore()
	savedState.State["uuid"] = &SavedState{
		Name: "testname",
		UUID: "uuid",
		State: map[devices.ID]devices.State{
			devID: {
				"on": false,
			},
		},
	}
	l := New(savedState, syncer)
	r := l.AddRule("test")
	r.Expression_ = `devices["node.id"].on == true`
	r.Enabled = true
	r.For_ = stypes.Duration(time.Millisecond * 40)
	r.Actions_ = []string{
		"uuid",
	}

	l.updateDevice(&devices.Device{
		ID: devID,
		State: devices.State{
			"on": true,
		},
	})
	l.EvaluateRules(context.Background())

	time.Sleep(30 * time.Millisecond)

	l.updateDevice(&devices.Device{
		ID: devID,
		State: devices.State{
			"on": false,
		},
	})
	l.EvaluateRules(context.Background())

	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, false, l.Rules[r.Uuid()].Active())

	l.Wait()
	assert.Equal(t, int64(0), syncer.Count())
}

// TestEvaluateRulesWithForFlapping asserts that we do not run the rule if it goes inactive before the "for" timeout even it it goes true again multiple times during the for.
func TestEvaluateRulesWithForFlapping(t *testing.T) {
	logrus.SetFormatter(&logrus.TextFormatter{TimestampFormat: time.RFC3339Nano, FullTimestamp: true})
	logrus.SetLevel(logrus.DebugLevel)
	syncer := NewMockSender()

	devID := devices.ID{Node: "node", ID: "id"}
	savedState := NewSavedStateStore()
	savedState.State["uuid"] = &SavedState{
		Name: "testname",
		UUID: "uuid",
		State: map[devices.ID]devices.State{
			devID: {
				"on": false,
			},
		},
	}
	l := New(savedState, syncer)
	r := l.AddRule("test")
	r.Expression_ = `devices["node.id"].on == true`
	r.Enabled = true
	r.For_ = stypes.Duration(time.Millisecond * 40)
	r.Actions_ = []string{
		"uuid",
	}

	l.updateDevice(&devices.Device{
		ID: devID,
		State: devices.State{
			"on": true,
		},
	})
	l.EvaluateRules(context.Background())

	time.Sleep(30 * time.Millisecond)
	l.updateDevice(&devices.Device{
		ID: devID,
		State: devices.State{
			"on": false,
		},
	})
	l.EvaluateRules(context.Background())

	now := time.Now()
	l.updateDevice(&devices.Device{
		ID: devID,
		State: devices.State{
			"on": true,
		},
	})
	l.EvaluateRules(context.Background())

	l.Wait()
	diff := time.Now().Sub(now)
	if diff < time.Millisecond*40 {
		t.Error("Expected to sleep in the action for at least 40ms but only slept: ", diff)
	}

	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, true, l.Rules[r.Uuid()].Active())

	l.Wait()
	assert.Equal(t, int64(1), syncer.Count())
}
