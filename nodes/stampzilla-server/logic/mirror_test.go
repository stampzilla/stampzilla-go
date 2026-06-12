package logic

import (
	"context"
	"testing"

	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/models/devices"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// seedDevice adds a device to a devices.List.
func seedDevice(devList *devices.List, node, id string, state devices.State) {
	devList.Add(&devices.Device{
		ID:    devices.ID{Node: node, ID: id},
		State: state,
	})
}

// TestEvalValue verifies evalValue with float, transform, and bool expressions.
func TestEvalValue(t *testing.T) {
	devs := devices.NewList()
	seedDevice(devs, "src", "1", devices.State{
		"brightness": float64(0.4),
		"on":         true,
	})

	rules := map[string]bool{}

	t.Run("float passthrough", func(t *testing.T) {
		r := &Rule{Expression_: `devices['src.1'].brightness`}
		got, err := r.EvalValue(devs, rules)
		require.NoError(t, err)
		assert.InDelta(t, float64(0.4), got, 1e-9)
	})

	t.Run("float transform", func(t *testing.T) {
		r := &Rule{Expression_: `devices['src.1'].brightness * 0.5`}
		got, err := r.EvalValue(devs, rules)
		require.NoError(t, err)
		assert.InDelta(t, float64(0.2), got, 1e-9)
	})

	t.Run("bool passthrough", func(t *testing.T) {
		r := &Rule{Expression_: `devices['src.1'].on`}
		got, err := r.EvalValue(devs, rules)
		require.NoError(t, err)
		assert.Equal(t, true, got)
	})

	t.Run("non-bool does not return ErrExpressionNotBool", func(t *testing.T) {
		// evalValue must NOT enforce bool — that is eval's job.
		r := &Rule{Expression_: `1 + 2`}
		_, err := r.EvalValue(devs, rules)
		assert.NoError(t, err)
	})
}

// TestRunMirrorBasic verifies that runMirror sends the source value to the target.
func TestRunMirrorBasic(t *testing.T) {
	syncer := NewMockSender()
	savedState := NewSavedStateStore()
	l := New(savedState, syncer)

	// Source device: srcNode.1 brightness = 0.4
	seedDevice(l.devices, "srcNode", "1", devices.State{"brightness": float64(0.4)})
	// Target device: dstNode.2 brightness = 0.0 (different, so send should happen)
	seedDevice(l.devices, "dstNode", "2", devices.State{"brightness": float64(0.0)})

	rule := &Rule{
		Uuid_:       "mirror-rule-1",
		Enabled:     true,
		Type_:       "mirror",
		Expression_: `devices['srcNode.1'].brightness`,
		Target_:     "dstNode.2",
		TargetKey_:  "brightness",
	}

	l.runMirror(rule)

	// One SendToID call should have happened.
	assert.Equal(t, int64(1), syncer.Count())
	// The mock sender adds the device to its Devices list.
	got := syncer.Devices.Get(devices.ID{Node: "dstNode", ID: "2"})
	require.NotNil(t, got)
	assert.InDelta(t, float64(0.4), got.State["brightness"], 1e-9)
}

// TestRunMirrorLoopBreaker verifies that runMirror does NOT send when the target
// already holds the same value (prevents feedback loops).
func TestRunMirrorLoopBreaker(t *testing.T) {
	syncer := NewMockSender()
	savedState := NewSavedStateStore()
	l := New(savedState, syncer)

	// Source and target both have brightness = 0.4 already.
	seedDevice(l.devices, "srcNode", "1", devices.State{"brightness": float64(0.4)})
	seedDevice(l.devices, "dstNode", "2", devices.State{"brightness": float64(0.4)})

	rule := &Rule{
		Uuid_:       "mirror-rule-2",
		Enabled:     true,
		Type_:       "mirror",
		Expression_: `devices['srcNode.1'].brightness`,
		Target_:     "dstNode.2",
		TargetKey_:  "brightness",
	}

	l.runMirror(rule)

	// No send should have happened.
	assert.Equal(t, int64(0), syncer.Count())
}

// TestRunMirrorLoopBreakerTypes documents that the == loop-breaker in runMirror is
// safe for every primitive type that can appear in device State (bool, float64, int,
// int64, string) and verifies correct send/no-send semantics.
//
// Note on int: CEL surfaces stored int values as int64 (CEL's integer type), so the
// loop-breaker compares int64 (from CEL) against int (in the target state). These are
// different dynamic types in the interface comparison — no panic, but the dedup never
// fires. The int/equal sub-test asserts this always-sends behaviour explicitly.
func TestRunMirrorLoopBreakerTypes(t *testing.T) {
	type testCase struct {
		srcVal              interface{}
		dstSameAsSrc        interface{}
		dstDiff             interface{}
		expectNoSendOnEqual bool // false for int: CEL int64 vs stored int → no match
	}
	tests := map[string]testCase{
		"bool":    {true, true, false, true},
		"float64": {float64(0.8), float64(0.8), float64(0.0), true},
		"string":  {"on", "on", "off", true},
		"int64":   {int64(5), int64(5), int64(9), true},
		// int: CEL converts the source int to int64; comparing int64(5)==int(5) is
		// false (different dynamic types), so the loop-breaker never suppresses the
		// send. No panic — both are comparable types.
		"int": {int(5), int(5), int(7), false},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name+"/equal", func(t *testing.T) {
			syncer := NewMockSender()
			l := New(NewSavedStateStore(), syncer)
			seedDevice(l.devices, "src", "1", devices.State{"val": tc.srcVal})
			seedDevice(l.devices, "dst", "2", devices.State{"val": tc.dstSameAsSrc})
			rule := &Rule{
				Uuid_:       "lb-eq-" + name,
				Enabled:     true,
				Type_:       "mirror",
				Expression_: `devices['src.1'].val`,
				Target_:     "dst.2",
				TargetKey_:  "val",
			}
			l.runMirror(rule)
			if tc.expectNoSendOnEqual {
				assert.Equal(t, int64(0), syncer.Count(), "loop-breaker should suppress send when value already matches")
			} else {
				assert.Equal(t, int64(1), syncer.Count(), "int always sends: CEL int64 vs stored int never match")
			}
		})

		t.Run(name+"/different", func(t *testing.T) {
			syncer := NewMockSender()
			l := New(NewSavedStateStore(), syncer)
			seedDevice(l.devices, "src", "1", devices.State{"val": tc.srcVal})
			seedDevice(l.devices, "dst", "2", devices.State{"val": tc.dstDiff})
			rule := &Rule{
				Uuid_:       "lb-diff-" + name,
				Enabled:     true,
				Type_:       "mirror",
				Expression_: `devices['src.1'].val`,
				Target_:     "dst.2",
				TargetKey_:  "val",
			}
			l.runMirror(rule)
			assert.Equal(t, int64(1), syncer.Count(), "send should occur when value differs")
		})
	}
}

// TestRunMirrorTransform verifies that a CEL transform is applied before forwarding.
func TestRunMirrorTransform(t *testing.T) {
	syncer := NewMockSender()
	savedState := NewSavedStateStore()
	l := New(savedState, syncer)

	seedDevice(l.devices, "srcNode", "1", devices.State{"brightness": float64(0.4)})
	seedDevice(l.devices, "dstNode", "2", devices.State{"brightness": float64(0.0)})

	rule := &Rule{
		Uuid_:       "mirror-rule-3",
		Enabled:     true,
		Type_:       "mirror",
		Expression_: `devices['srcNode.1'].brightness * 0.5`,
		Target_:     "dstNode.2",
		TargetKey_:  "brightness",
	}

	l.runMirror(rule)

	assert.Equal(t, int64(1), syncer.Count())
	got := syncer.Devices.Get(devices.ID{Node: "dstNode", ID: "2"})
	require.NotNil(t, got)
	assert.InDelta(t, float64(0.2), got.State["brightness"], 1e-9)
}

// TestRunMirrorUnknownTarget verifies that runMirror still sends when the target
// is not yet known in the local device store (first-time setup).
func TestRunMirrorUnknownTarget(t *testing.T) {
	syncer := NewMockSender()
	savedState := NewSavedStateStore()
	l := New(savedState, syncer)

	// Source device exists, but target is NOT in l.devices yet.
	seedDevice(l.devices, "srcNode", "1", devices.State{"brightness": float64(0.7)})

	rule := &Rule{
		Uuid_:       "mirror-rule-4",
		Enabled:     true,
		Type_:       "mirror",
		Expression_: `devices['srcNode.1'].brightness`,
		Target_:     "dstNode.2",
		TargetKey_:  "brightness",
	}

	l.runMirror(rule)

	// Should send even though target was unknown.
	assert.Equal(t, int64(1), syncer.Count())
}

// TestRunMirrorConfigErrors verifies that misconfigured mirror rules report errors
// via onReportState and never call SendToID.
func TestRunMirrorConfigErrors(t *testing.T) {
	tests := []struct {
		name      string
		rule      *Rule
		wantError string
	}{
		{
			name: "empty target reports error",
			rule: &Rule{
				Uuid_:       "e1",
				Enabled:     true,
				Type_:       "mirror",
				Expression_: `devices['src.1'].brightness`,
				Target_:     "",
				TargetKey_:  "brightness",
			},
			wantError: "invalid target",
		},
		{
			name: "empty targetKey reports error",
			rule: &Rule{
				Uuid_:       "e2",
				Enabled:     true,
				Type_:       "mirror",
				Expression_: `devices['src.1'].brightness`,
				Target_:     "dst.2",
				TargetKey_:  "",
			},
			wantError: "targetKey is not configured",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			syncer := NewMockSender()
			savedState := NewSavedStateStore()
			l := New(savedState, syncer)
			seedDevice(l.devices, "src", "1", devices.State{"brightness": float64(0.5)})

			var gotError string
			l.OnReportState(func(_ string, state devices.State) {
				if e, ok := state["error"].(string); ok && e != "" {
					gotError = e
				}
			})

			l.runMirror(tt.rule)

			assert.Equal(t, int64(0), syncer.Count(), "SendToID should not be called on config error")
			assert.Contains(t, gotError, tt.wantError)
		})
	}
}

// TestEvaluateRulesMirrorIntegration verifies that EvaluateRules dispatches mirror
// rules without entering the bool edge-transition path.
func TestEvaluateRulesMirrorIntegration(t *testing.T) {
	syncer := NewMockSender()
	savedState := NewSavedStateStore()
	l := New(savedState, syncer)

	// Populate l.devices directly (same package).
	l.updateDevice(&devices.Device{
		ID:    devices.ID{Node: "srcNode", ID: "1"},
		State: devices.State{"brightness": float64(0.6)},
	})
	l.updateDevice(&devices.Device{
		ID:    devices.ID{Node: "dstNode", ID: "2"},
		State: devices.State{"brightness": float64(0.0)},
	})

	l.Rules["m1"] = &Rule{
		Uuid_:       "m1",
		Enabled:     true,
		Type_:       "mirror",
		Expression_: `devices['srcNode.1'].brightness`,
		Target_:     "dstNode.2",
		TargetKey_:  "brightness",
	}

	l.EvaluateRules(context.Background())

	assert.Equal(t, int64(1), syncer.Count())
	got := syncer.Devices.Get(devices.ID{Node: "dstNode", ID: "2"})
	require.NotNil(t, got)
	assert.InDelta(t, float64(0.6), got.State["brightness"], 1e-9)
}
