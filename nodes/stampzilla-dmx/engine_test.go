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

// fillOnceConfig uses 6 fixtures (not testConfig's 3): count-1 needs to be
// large enough to distinguish "interrupted mid-animation" from "already at
// the extreme end", which 3 fixtures can't reliably do.
const fillOnceConfig = `{
	"fixtures": {
		"f1": {"profile": "dimmer", "address": 1},
		"f2": {"profile": "dimmer", "address": 2},
		"f3": {"profile": "dimmer", "address": 3},
		"f4": {"profile": "dimmer", "address": 4},
		"f5": {"profile": "dimmer", "address": 5},
		"f6": {"profile": "dimmer", "address": 6}
	},
	"groups": {
		"g1": {"fixtures": ["f1", "f2", "f3", "f4", "f5", "f6"], "pattern": "fillonce", "interval": "100ms"}
	}
}`

func litCount(frame []byte, n int) int {
	lit := 0
	for _, b := range frame[:n] {
		if b != 0 {
			lit++
		}
	}
	return lit
}

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

// TestEngineFillOnceStartsSettledWhenOff is the regression test for the
// startup-flash bug: a fresh group is off from config load (no
// setGroupState call at all here), and must render fully dark on its very
// first frame rather than showing step 0 of a closing animation it never
// actually played.
func TestEngineFillOnceStartsSettledWhenOff(t *testing.T) {
	e, out := newTestEngine(t, fillOnceConfig)
	e.renderFrame()
	if lit := litCount(out.frames[0], 6); lit != 0 {
		t.Errorf("lit = %d, want 0 - a fresh off group must start settled dark, not mid-drain", lit)
	}
}

// TestEngineFillOnceDrainsWhileOff proves the rendersWhileOff bypass in
// renderFrame actually engages: an off fillonce group is not simply skipped.
func TestEngineFillOnceDrainsWhileOff(t *testing.T) {
	e, out := newTestEngine(t, fillOnceConfig)
	setGroupState(t, e, "g1", false, 1, "fillonce", 0)
	e.renderFrame()
	frame := out.frames[0]
	if lit := litCount(frame, 6); lit != 5 {
		t.Errorf("lit = %d, want 5 (count-1) at the start of the drain", lit)
	}
	if frame[5] != 0 {
		t.Error("the last fixture (index 5) should be the first to turn off when draining")
	}
}

func TestEngineFillOnceOffTerminalIsBlack(t *testing.T) {
	e, out := newTestEngine(t, fillOnceConfig)

	setGroupState(t, e, "g1", false, 1, "fillonce", 6*100*time.Millisecond) // count*interval
	e.renderFrame()
	if lit := litCount(out.frames[0], 6); lit != 0 {
		t.Errorf("lit = %d, want 0 at the end of the drain", lit)
	}

	setGroupState(t, e, "g1", false, 1, "fillonce", 60*100*time.Millisecond) // well past the end
	e.renderFrame()
	if lit := litCount(out.frames[1], 6); lit != 0 {
		t.Errorf("lit = %d, want 0 - a drained group must hold dark, not re-fill", lit)
	}
}

// TestEngineSwitchToFillOnceWhileOffStaysDark is the regression test for the
// switch-while-off flash: a group already off under a different pattern
// must settle dark immediately when switched to fillonce, not animate a
// closing sequence that was never playing.
func TestEngineSwitchToFillOnceWhileOffStaysDark(t *testing.T) {
	e, out := newTestEngine(t, fillOnceConfig)
	setGroupState(t, e, "g1", false, 1, "static", 0)

	if err := e.onRequestStateChange(devices.State{"pattern": "fillonce"}, &devices.Device{ID: devices.ID{ID: "g1"}}); err != nil {
		t.Fatalf("onRequestStateChange() error = %v", err)
	}

	e.renderFrame()
	if lit := litCount(out.frames[0], 6); lit != 0 {
		t.Errorf("lit = %d, want 0 - switching to fillonce while off must not flash", lit)
	}
}

// TestEngineFillOnceInterruptedFillCarriesProgress is the regression test
// for the interrupted-toggle jump: turning a group off partway through
// filling in must not first flash the fixtures the fill hadn't reached yet
// before draining - the drain must pick up from the currently-visible lit
// count.
func TestEngineFillOnceInterruptedFillCarriesProgress(t *testing.T) {
	e, out := newTestEngine(t, fillOnceConfig)
	setGroupState(t, e, "g1", true, 1, "fillonce", 2*100*time.Millisecond) // step 2 -> 3 lit
	e.renderFrame()
	litBefore := litCount(out.frames[0], 6)
	if litBefore != 3 {
		t.Fatalf("setup: lit = %d, want 3 before toggling off", litBefore)
	}

	if err := e.onRequestStateChange(devices.State{"on": false}, &devices.Device{ID: devices.ID{ID: "g1"}}); err != nil {
		t.Fatalf("onRequestStateChange() error = %v", err)
	}
	e.renderFrame()
	litAfter := litCount(out.frames[1], 6)
	if litAfter > litBefore {
		t.Errorf("lit jumped from %d to %d after toggling off - interrupted fill must not flash extra fixtures before draining", litBefore, litAfter)
	}
}

// TestEngineFillOnceInterruptedDrainCarriesProgress mirrors
// TestEngineFillOnceInterruptedFillCarriesProgress in the other direction.
func TestEngineFillOnceInterruptedDrainCarriesProgress(t *testing.T) {
	e, out := newTestEngine(t, fillOnceConfig)
	setGroupState(t, e, "g1", false, 1, "fillonce", 3*100*time.Millisecond) // closing step 3 -> 2 lit
	e.renderFrame()
	litBefore := litCount(out.frames[0], 6)
	if litBefore != 2 {
		t.Fatalf("setup: lit = %d, want 2 before toggling on", litBefore)
	}

	if err := e.onRequestStateChange(devices.State{"on": true}, &devices.Device{ID: devices.ID{ID: "g1"}}); err != nil {
		t.Fatalf("onRequestStateChange() error = %v", err)
	}
	e.renderFrame()
	litAfter := litCount(out.frames[1], 6)
	if litAfter < litBefore {
		t.Errorf("lit dropped from %d to %d after toggling on - interrupted drain must not flash dark before filling back up", litBefore, litAfter)
	}
}

// TestEngineFillOnceRedundantOffDoesNotRestartDrain guards the
// sawOn/oldOn-based flip detection in onRequestStateChange: resending the
// same "on" value must not touch startedAt.
func TestEngineFillOnceRedundantOffDoesNotRestartDrain(t *testing.T) {
	e, _ := newTestEngine(t, fillOnceConfig)
	setGroupState(t, e, "g1", false, 1, "fillonce", 2*100*time.Millisecond)

	e.mu.Lock()
	before := e.states["g1"].startedAt
	e.mu.Unlock()

	if err := e.onRequestStateChange(devices.State{"on": false}, &devices.Device{ID: devices.ID{ID: "g1"}}); err != nil {
		t.Fatalf("onRequestStateChange() error = %v", err)
	}

	e.mu.Lock()
	after := e.states["g1"].startedAt
	e.mu.Unlock()

	if !before.Equal(after) {
		t.Errorf("startedAt moved from %v to %v on a redundant {on:false} - should be a no-op", before, after)
	}
}

// TestEngineNonLatchingPatternStillInstantOff confirms the other 10
// patterns are unaffected by rendersWhileOff - TestEngineOffBlacksOutUniverse
// only covers "static"; "fill" is the pattern someone would most plausibly
// mistake for wanting the new behavior, so it's worth its own check.
func TestEngineNonLatchingPatternStillInstantOff(t *testing.T) {
	e, out := newTestEngine(t, testConfig)
	setGroupState(t, e, "g1", false, 1, "fill", 2*100*time.Millisecond)
	e.renderFrame()
	for i, b := range out.frames[0] {
		if b != 0 {
			t.Errorf("byte %d = %d, want 0 (fill is not a rendersWhileOff pattern, must be instant-off)", i, b)
		}
	}
}

func TestEngineFillOnceBrightnessAppliesWhileDraining(t *testing.T) {
	e, out := newTestEngine(t, fillOnceConfig)
	setGroupState(t, e, "g1", false, 0.5, "fillonce", 0) // step 0 closing -> fixture 0 lit
	e.renderFrame()
	if got := out.frames[0][0]; got != 128 {
		t.Errorf("dimmer channel = %d, want 128 (255 scaled by 0.5 brightness) while draining", got)
	}
}

// TestEngineFillOnceOffKeepsProfileStaticChannels pins a documented
// divergence from every other pattern's off behavior: because an off
// fillonce group keeps calling writeFixture (at intensity 0) instead of
// being skipped, a static value on a color/dimmer role channel is
// overwritten to 0, but a static value on any other role (like "mode" here)
// is left untouched.
func TestEngineFillOnceOffKeepsProfileStaticChannels(t *testing.T) {
	config := `{
		"profiles": {"withmode": {"channels": ["mode", "dimmer"], "static": {"mode": 200}}},
		"fixtures": {"f1": {"profile": "withmode", "address": 1}},
		"groups": {"g1": {"fixtures": ["f1"], "pattern": "fillonce", "interval": "100ms"}}
	}`
	e, out := newTestEngine(t, config)
	setGroupState(t, e, "g1", false, 1, "fillonce", 10*time.Second) // fully drained
	e.renderFrame()
	frame := out.frames[0]
	if frame[0] != 200 {
		t.Errorf("mode channel = %d, want 200 (static, unaffected by fillonce's off-drain)", frame[0])
	}
	if frame[1] != 0 {
		t.Errorf("dimmer channel = %d, want 0 (fillonce's off-drain zeroes color/dimmer channels)", frame[1])
	}
}

// TestEngineFillOnceStopsClobberingSharedFixturesAfterSettling is the
// regression test for a real-world bug: several groups (alternative
// effects/"scenes") declared over the same physical fixtures, toggled one
// at a time. Group keys are rendered in sortedKeys order, so a settled, off
// fillonce group whose key sorts after an active group's key was
// overwriting that group's output with zeros every frame, forever, because
// renderFrame never stopped rendering it once its animation had finished.
// The key names here ("a_on"/"b_fillonce") are chosen so alphabetical order
// reproduces that exact ordering.
func TestEngineFillOnceStopsClobberingSharedFixturesAfterSettling(t *testing.T) {
	config := `{
		"fixtures": {
			"f1": {"profile": "dimmer", "address": 1},
			"f2": {"profile": "dimmer", "address": 2}
		},
		"groups": {
			"a_on": {"fixtures": ["f1", "f2"], "pattern": "static", "colors": ["#ffffff"], "interval": "100ms"},
			"b_fillonce": {"fixtures": ["f1", "f2"], "pattern": "fillonce", "interval": "100ms"}
		}
	}`
	e, out := newTestEngine(t, config)
	setGroupState(t, e, "a_on", true, 1, "static", 0)
	setGroupState(t, e, "b_fillonce", false, 1, "fillonce", 10*time.Second) // long settled

	e.renderFrame()
	frame := out.frames[0]
	if frame[0] == 0 || frame[1] == 0 {
		t.Errorf("frame = %v, want both fixtures lit from the 'a_on' static group - a settled, off fillonce group sorting after it must not keep clobbering shared fixtures with zeros", frame)
	}
}
