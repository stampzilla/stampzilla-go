package logic

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/models/devices"
	"github.com/stretchr/testify/assert"
)

func TestEval(t *testing.T) {

	tests := []struct {
		State       devices.State
		Expression  string
		Expected    bool
		ExpectedErr error
	}{
		{
			State: devices.State{
				"on":          true,
				"temperature": 21.0,
			},
			Expression: `devices["node.id"].on == true && devices["node.id"].temperature > 20.0 `,
			Expected:   true,
		},
		{
			State: devices.State{
				"on": true,
			},
			Expression: `devices["node.id"].on == false`,
			Expected:   false,
		},
		{
			State: devices.State{
				"on": true,
			},
			Expression:  `1+2`,
			ExpectedErr: ErrExpressionNotBool,
		},
		{
			State: devices.State{
				"on": true,
			},
			Expression: `rules["rule1"] == true `,
			Expected:   true,
		},
		{
			State: devices.State{
				"on": true,
			},
			Expression:  `rules["rule"] == true `,
			Expected:    false,
			ExpectedErr: fmt.Errorf("no such key: 'rule'"),
		},
	}

	for _, v := range tests {
		t.Run(v.Expression, func(t *testing.T) {
			rules := make(map[string]bool)
			rules["rule1"] = true
			devs := devices.NewList()
			devs.Add(&devices.Device{
				ID: devices.ID{
					Node: "node",
					ID:   "id",
				},
				State: v.State,
			})
			r := &Rule{
				Expression_: v.Expression,
			}
			result, err := r.Eval(devs, rules)
			assert.Equal(t, v.ExpectedErr, err)
			assert.Equal(t, v.Expected, result)
		})

	}

}

func TestRunActions(t *testing.T) {
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

	syncer.Devices.Add(&devices.Device{
		ID: devices.ID{
			Node: "node",
			ID:   "id",
		},
		State: devices.State{
			"on": true,
		},
	})

	r := &Rule{
		Actions_: []string{
			"400ms",
			"uuid",
		},
	}

	now := time.Now()
	r.Run(savedState, syncer)
	if time.Now().Sub(now) < time.Millisecond*200 {
		t.Error("Expected to sleep in the action for at least 200ms")
	}

	//t.Log(store.Devices.Get("node", "id"))
	assert.Equal(t, false, syncer.Devices.Get(devices.ID{"node", "id"}).State["on"])
	assert.Equal(t, true, syncer.Devices.Get(devices.ID{"node", "id"}).State["a"])
	assert.Equal(t, true, syncer.Devices.Get(devices.ID{"node", "id"}).State["b"])
}

func TestRunActionsCancelSleep(t *testing.T) {
	var logBuf strings.Builder
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetOutput(&logBuf)

	syncer := NewMockSender()
	savedState := NewSavedStateStore()

	r := &Rule{
		Actions_: []string{
			"100ms",
			"100ms",
		},
	}

	go func() {
		time.Sleep(110 * time.Millisecond)
		r.Cancel()
	}()

	now := time.Now()
	r.Run(savedState, syncer)
	dur := time.Now().Sub(now)
	if dur < time.Millisecond*110 {
		t.Error("Expected to sleep in the action for at least 200ms slept: ", dur)
	}
	assert.Contains(t, logBuf.String(), "stopping action 1 due to cancel")
}
func BenchmarkEval(b *testing.B) {
	devs := devices.NewList()
	devs.Add(&devices.Device{
		ID: devices.ID{
			Node: "node",
			ID:   "id",
		},
		State: devices.State{
			"on": true,
		},
	})
	r := &Rule{
		Expression_: `devices["node.id"].on == true `,
	}

	rules := make(map[string]bool)
	for i := 0; i < b.N; i++ {
		result, err := r.Eval(devs, rules)
		assert.NoError(b, err)
		assert.Equal(b, true, result)
	}
}
