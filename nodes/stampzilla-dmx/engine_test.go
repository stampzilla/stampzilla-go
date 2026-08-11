package main

import (
	"testing"
	"time"

	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/models/devices"
	"github.com/stampzilla/stampzilla-go/v2/pkg/node"
)

// fakeOutput captures every frame sent to it instead of writing to hardware.
type fakeOutput struct {
	frames [][]byte
	closed bool
}

func (f *fakeOutput) Send(channels []byte) error {
	cp := make([]byte, len(channels))
	copy(cp, channels)
	f.frames = append(f.frames, cp)
	return nil
}

func (f *fakeOutput) Close() error {
	f.closed = true
	return nil
}

// newTestEngine builds an engine loaded with configJSON, with ensureGroupDevice
// and setDeviceOnline stubbed out. AddOrUpdate/SyncDevice block on the node's
// unbuffered sendUpdate channel unless a real Connect()'ed node is running its
// syncWorker (see the same pattern in stampzilla-denonavr's tests), so a plain
// node.NewWithClient(nil) cannot be driven through those calls directly.
func newTestEngine(t *testing.T, configJSON string) (*engine, *fakeOutput) {
	t.Helper()

	origEnsure := ensureGroupDevice
	origOnline := setDeviceOnline
	t.Cleanup(func() {
		ensureGroupDevice = origEnsure
		setDeviceOnline = origOnline
	})
	ensureGroupDevice = func(*node.Node, string, string, string) {}
	setDeviceOnline = func(*node.Node, string, bool) {}

	n := node.NewWithClient(nil)
	e := newEngine(n)
	if err := e.updatedConfig([]byte(configJSON)); err != nil {
		t.Fatalf("updatedConfig() error = %v", err)
	}

	out := &fakeOutput{}
	e.setOutput(out)

	return e, out
}

// setGroupState overrides a group's runtime state directly, backdating
// startedAt by elapsed so renderFrame() sees a specific, deterministic
// step/phase.
func setGroupState(t *testing.T, e *engine, id string, on bool, brightness float64, pattern string, elapsed time.Duration) {
	t.Helper()
	e.mu.Lock()
	defer e.mu.Unlock()
	st, ok := e.states[id]
	if !ok {
		t.Fatalf("no such group %q", id)
	}
	st.on = on
	st.brightness = brightness
	if pattern != "" {
		st.pattern = pattern
	}
	st.startedAt = time.Now().Add(-elapsed)
}

const testConfig = `{
	"fixtures": {
		"f1": {"profile": "rgb", "address": 1},
		"f2": {"profile": "rgb", "address": 4},
		"f3": {"profile": "rgb", "address": 7}
	},
	"groups": {
		"g1": {"fixtures": ["f1", "f2", "f3"], "pattern": "static", "colors": ["#ff0000"], "interval": "100ms"}
	}
}`

func TestEngineOffBlacksOutUniverse(t *testing.T) {
	e, out := newTestEngine(t, testConfig)
	setGroupState(t, e, "g1", false, 1, "", 0)
	e.renderFrame()
	if len(out.frames) != 1 {
		t.Fatalf("frames sent = %d, want 1", len(out.frames))
	}
	for i, b := range out.frames[0] {
		if b != 0 {
			t.Errorf("byte %d = %d, want 0 (group off)", i, b)
		}
	}
}

func TestEngineStaticPattern(t *testing.T) {
	e, out := newTestEngine(t, testConfig)
	setGroupState(t, e, "g1", true, 1, "static", 0)
	e.renderFrame()
	frame := out.frames[0]
	want := []byte{255, 0, 0} // red, from the group's configured color
	for _, offset := range []int{0, 3, 6} {
		got := frame[offset : offset+3]
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("fixture at offset %d channel %d = %d, want %d", offset, i, got[i], want[i])
			}
		}
	}
}

func TestEngineBrightnessScalesOutput(t *testing.T) {
	e, out := newTestEngine(t, testConfig)
	setGroupState(t, e, "g1", true, 0.5, "static", 0)
	e.renderFrame()
	if got := out.frames[0][0]; got != 128 {
		t.Errorf("red channel = %d, want 128 (255 scaled by 0.5 brightness)", got)
	}
}

func TestEngineChaseAdvancesWithSteps(t *testing.T) {
	e, out := newTestEngine(t, testConfig)

	setGroupState(t, e, "g1", true, 1, "chase", 0)
	e.renderFrame()
	f0 := out.frames[0]
	if f0[0] == 0 || f0[3] != 0 || f0[6] != 0 {
		t.Fatalf("chase step 0: want only fixture 0 lit, got %v", f0)
	}

	setGroupState(t, e, "g1", true, 1, "chase", 100*time.Millisecond)
	e.renderFrame()
	f1 := out.frames[1]
	if f1[3] == 0 || f1[0] != 0 || f1[6] != 0 {
		t.Fatalf("chase step 1: want only fixture 1 lit, got %v", f1)
	}
}

func TestEngineReverseMirrorsFixtureOrder(t *testing.T) {
	e, out := newTestEngine(t, testConfig)
	setGroupState(t, e, "g1", true, 1, "chase", 0)

	e.mu.Lock()
	e.cfg.groups["g1"].reverse = true
	e.mu.Unlock()

	e.renderFrame()
	f := out.frames[0]
	if f[6] == 0 || f[0] != 0 || f[3] != 0 {
		t.Fatalf("reversed chase step 0: want the last fixture lit, got %v", f)
	}
}

func TestEngineProfileStaticChannelsAlwaysApplied(t *testing.T) {
	config := `{
		"profiles": {"withmode": {"channels": ["mode", "dimmer"], "static": {"mode": 200}}},
		"fixtures": {"f1": {"profile": "withmode", "address": 1}},
		"groups": {"g1": {"fixtures": ["f1"], "pattern": "off"}}
	}`
	e, out := newTestEngine(t, config)
	setGroupState(t, e, "g1", false, 1, "", 0)
	e.renderFrame()
	if got := out.frames[0][0]; got != 200 {
		t.Errorf("mode channel = %d, want 200 (static, applied even with the group off)", got)
	}
}

func TestEngineUnknownGroupRejected(t *testing.T) {
	e, _ := newTestEngine(t, testConfig)
	err := e.onRequestStateChange(devices.State{"on": true}, &devices.Device{ID: devices.ID{ID: "nope"}})
	if err == nil {
		t.Fatal("onRequestStateChange() error = nil, want error for unknown group")
	}
}

func TestEngineUnknownPatternRejectedWithoutMutating(t *testing.T) {
	e, _ := newTestEngine(t, testConfig)
	err := e.onRequestStateChange(devices.State{"pattern": "nope"}, &devices.Device{ID: devices.ID{ID: "g1"}})
	if err == nil {
		t.Fatal("onRequestStateChange() error = nil, want error for unknown pattern")
	}
	e.mu.Lock()
	got := e.states["g1"].pattern
	e.mu.Unlock()
	if got == "nope" {
		t.Errorf("pattern was applied despite failing validation")
	}
}

func TestEngineRequestStateChangeAppliesOnAndBrightness(t *testing.T) {
	e, _ := newTestEngine(t, testConfig)
	err := e.onRequestStateChange(devices.State{"on": true, "brightness": 0.25}, &devices.Device{ID: devices.ID{ID: "g1"}})
	if err != nil {
		t.Fatalf("onRequestStateChange() error = %v", err)
	}
	e.mu.Lock()
	st := *e.states["g1"]
	e.mu.Unlock()
	if !st.on || st.brightness != 0.25 {
		t.Errorf("state = %+v, want on=true brightness=0.25", st)
	}
}
