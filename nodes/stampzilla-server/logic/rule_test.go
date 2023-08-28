package logic

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/models/devices"
	"github.com/stretchr/testify/assert"
)

func TestSetUUID(t *testing.T) {
	r := &Rule{
		Uuid_: "a",
	}

	assert.Equal(t, r.Uuid(), "a")

	r.SetUuid("b")

	assert.Equal(t, r.Uuid(), "b")
}

func TestConditions(t *testing.T) {
	c := map[string]bool{
		"a": true,
		"b": false,
	}
	r := &Rule{
		Conditions_: c,
	}

	assert.Equal(t, r.Conditions(), c)
}

func TestType(t *testing.T) {
	r := &Rule{
		Type_: "type1",
	}

	assert.Equal(t, r.Type(), "type1")
}

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
			ExpectedErr: fmt.Errorf("no such key: rule"),
		},
		{
			State: devices.State{
				"on": true,
			},
			Expression: `daily("00:00", "23:59") && devices["node.id"].on == true`,
			Expected:   true,
		},
		{
			State: devices.State{
				"on": true,
			},
			Expression: `daily("00:00", "00:01") && devices["node.id"].on == true`,
			Expected:   false, // Yes this will fail between 00:00 and 00:01. TODO implement some nowFunc which we can overwride in tests
		},
		/*
			{
				State: devices.State{
					"on": true,
				},
				Expression: `daily("17:00", "22:00") && devices["node.id"].on == true`,
				Expected:   false, // Yes this will fail between 00:00 and 00:01. TODO implement some nowFunc which we can overwride in tests
			},
		*/
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
			if err != nil {
				t.Log("error was: ", err)
			}
			if v.ExpectedErr != nil {

				assert.Equal(t, v.ExpectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
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
			{Node: "node", ID: "id"}: {
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
			"50ms",
			"uuid",
		},
		Destinations_: []string{
			"a",
			"b",
		},
	}

	now := time.Now()
	cnt := 0
	triggerDestination := func(string, string) error {
		cnt = cnt + 1
		return nil
	}

	r.Run(savedState, syncer, triggerDestination)
	if time.Now().Sub(now) < time.Millisecond*25 {
		t.Error("Expected to sleep in the action for at least 25ms")
	}

	// t.Log(store.Devices.Get("node", "id"))
	assert.Equal(t, false, syncer.Devices.Get(devices.ID{Node: "node", ID: "id"}).State["on"])
	assert.Equal(t, true, syncer.Devices.Get(devices.ID{Node: "node", ID: "id"}).State["a"])
	assert.Equal(t, true, syncer.Devices.Get(devices.ID{Node: "node", ID: "id"}).State["b"])
	assert.Equal(t, 2, cnt)
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
		Destinations_: []string{
			"a",
			"b",
			"c",
		},
	}

	go func() {
		time.Sleep(110 * time.Millisecond)
		r.Cancel()
	}()

	now := time.Now()
	cnt := 0
	triggerDestination := func(string, string) error {
		cnt = cnt + 1
		return nil
	}

	r.Run(savedState, syncer, triggerDestination)
	dur := time.Now().Sub(now)
	if dur < time.Millisecond*110 {
		t.Error("Expected to sleep in the action for at least 200ms slept: ", dur)
	}
	assert.Contains(t, logBuf.String(), "stopping action 1 due to cancel")
	assert.Equal(t, 0, cnt)
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
