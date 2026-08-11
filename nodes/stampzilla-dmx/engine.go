package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/models/devices"
	"github.com/stampzilla/stampzilla-go/v2/pkg/node"
)

const defaultFrameInterval = time.Second / defaultFPS

// groupState is the mutable runtime state of a group device. It survives
// config reloads for groups that keep existing (so an operator toggling a
// group on isn't reset by an unrelated config edit).
type groupState struct {
	on         bool
	brightness float64
	pattern    string
	startedAt  time.Time
}

// engine owns the resolved config, per-group runtime state, the universe
// buffer and the DMX output. All mutable state is guarded by mu, except the
// output itself which is accessed through the atomic outputBox so the frame
// loop never blocks on the connection supervisor.
type engine struct {
	node *node.Node

	mu     sync.Mutex
	cfg    *resolvedConfig
	states map[string]*groupState

	outputBox    atomic.Value // holds outputBox
	outputBroken atomic.Bool

	rateSignal   chan struct{}
	outputSignal chan struct{}
}

type outputBoxValue struct{ out dmxOutput }

func newEngine(n *node.Node) *engine {
	e := &engine{
		node:         n,
		states:       make(map[string]*groupState),
		rateSignal:   make(chan struct{}, 1),
		outputSignal: make(chan struct{}, 1),
	}
	e.outputBox.Store(outputBoxValue{logOutput{}})
	return e
}

func (e *engine) output() dmxOutput {
	return e.outputBox.Load().(outputBoxValue).out
}

func (e *engine) setOutput(o dmxOutput) {
	e.outputBox.Store(outputBoxValue{o})
}

func notify(ch chan struct{}) {
	select {
	case ch <- struct{}{}:
	default:
	}
}

// updatedConfig is registered as the node's OnConfig callback.
func (e *engine) updatedConfig(data json.RawMessage) error {
	cfg, err := loadConfig(data)
	if err != nil {
		return fmt.Errorf("dmx: bad config: %w", err)
	}

	e.mu.Lock()
	old := e.cfg
	e.cfg = cfg

	newStates := make(map[string]*groupState, len(cfg.groups))
	for key, g := range cfg.groups {
		if st, ok := e.states[key]; ok {
			newStates[key] = st
			continue
		}
		newStates[key] = &groupState{brightness: 1, pattern: g.pattern, startedAt: time.Now()}
	}
	e.states = newStates
	e.mu.Unlock()

	e.reconcileDevices(old, cfg)

	notify(e.rateSignal)
	notify(e.outputSignal)

	return nil
}

func (e *engine) reconcileDevices(old, cfg *resolvedConfig) {
	for _, key := range sortedKeys(cfg.groups) {
		g := cfg.groups[key]
		ensureGroupDevice(e.node, key, g.name, g.pattern)
	}
	if old == nil {
		return
	}
	for _, key := range sortedKeys(old.groups) {
		if _, ok := cfg.groups[key]; !ok {
			setDeviceOnline(e.node, key, false)
		}
	}
}

// ensureGroupDevice and setDeviceOnline are indirections so tests can stub
// them: AddOrUpdate/SyncDevice block on the node's unbuffered sendUpdate
// channel unless a real Connect()'ed node is running its syncWorker.
var ensureGroupDevice = func(n *node.Node, id, name, pattern string) {
	if name == "" {
		name = id
	}
	if dev := n.GetDevice(id); dev != nil {
		dev.Lock()
		changed := dev.Name != name
		if changed {
			dev.Name = name
		}
		dev.Unlock()
		if changed {
			n.AddOrUpdate(dev)
		}
		return
	}
	n.AddOrUpdate(&devices.Device{
		Type:   "light",
		ID:     devices.ID{ID: id},
		Name:   name,
		Online: false,
		Traits: []string{"OnOff", "Brightness"},
		State:  devices.State{"on": false, "brightness": 1.0, "pattern": pattern},
	})
}

var setDeviceOnline = func(n *node.Node, id string, online bool) {
	n.SetDeviceOnline(id, online)
}

// onRequestStateChange is registered as the node's OnRequestStateChange
// callback. It validates the requested change before mutating anything,
// since returning an error rejects the whole diff.
func (e *engine) onRequestStateChange(state devices.State, device *devices.Device) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	st, ok := e.states[device.ID.ID]
	if !ok {
		return fmt.Errorf("dmx: unknown group %s", device.ID.ID)
	}

	var newPattern string
	state.String("pattern", func(p string) { newPattern = p })
	if newPattern != "" {
		if _, ok := patterns[newPattern]; !ok {
			return fmt.Errorf("dmx: unknown pattern %q", newPattern)
		}
	}

	state.Bool("on", func(on bool) { st.on = on })
	state.Float("brightness", func(b float64) { st.brightness = clamp01(b) })
	if newPattern != "" {
		st.pattern = newPattern
		st.startedAt = time.Now() // restart the pattern from the beginning
	}

	return nil
}

func (e *engine) setGroupsOnline(online bool) {
	e.mu.Lock()
	keys := sortedKeys(e.states)
	e.mu.Unlock()
	for _, k := range keys {
		setDeviceOnline(e.node, k, online)
	}
}

// run starts the connection supervisor and the frame loop. It blocks until
// ctx is cancelled.
func (e *engine) run(ctx context.Context) {
	go e.connectionSupervisor(ctx)

	ticker := time.NewTicker(defaultFrameInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-e.rateSignal:
			e.mu.Lock()
			cfg := e.cfg
			e.mu.Unlock()
			if cfg != nil {
				ticker.Reset(time.Second / time.Duration(cfg.fps))
			}
		case <-ticker.C:
			e.renderFrame()
		}
	}
}

// close shuts down the current output. Call after the node has stopped.
func (e *engine) close() {
	_ = e.output().Close()
}

// connectionSupervisor owns opening, closing and retrying the DMX output. It
// runs independently from the frame loop so a stuck reconnect never blocks
// frame rendering.
func (e *engine) connectionSupervisor(ctx context.Context) {
	var openPort string
	var connected bool

	reconnect := func() {
		e.mu.Lock()
		cfg := e.cfg
		e.mu.Unlock()
		if cfg == nil {
			return
		}

		if cfg.port == "" {
			if connected {
				_ = e.output().Close()
			}
			e.setOutput(logOutput{})
			e.outputBroken.Store(false)
			e.setGroupsOnline(true)
			openPort, connected = "", false
			return
		}

		needsReopen := !connected || cfg.port != openPort || e.outputBroken.Load()
		if !needsReopen {
			return
		}

		_ = e.output().Close()
		o, err := openOpenDMXOutput(cfg.port)
		openPort = cfg.port
		e.outputBroken.Store(false)
		if err != nil {
			logrus.Errorf("dmx: open %s: %s", cfg.port, err)
			e.setOutput(logOutput{})
			e.setGroupsOnline(false)
			connected = false
			return
		}

		e.setOutput(o)
		e.setGroupsOnline(true)
		connected = true
	}

	reconnect()

	retry := time.NewTicker(5 * time.Second)
	defer retry.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-e.outputSignal:
			reconnect()
		case <-retry.C:
			reconnect()
		}
	}
}

// renderFrame computes one universe frame from the current config and group
// runtime state, and sends it to the output. It calls no node.* method:
// AddOrUpdate/SyncDevice/UpdateState would either block or flood the server
// at frame rate, so device state is only ever pushed from
// onRequestStateChange.
func (e *engine) renderFrame() {
	e.mu.Lock()
	cfg := e.cfg
	if cfg == nil {
		e.mu.Unlock()
		return
	}

	buf := make([]byte, cfg.universeSize)

	// Static channels are applied for every declared fixture, regardless of
	// group membership or on/off state.
	for _, fx := range cfg.fixtures {
		start, _ := fx.span()
		for offset, v := range fx.profile.static {
			pos := start + offset
			if pos >= 0 && pos < len(buf) {
				buf[pos] = v
			}
		}
	}

	for _, key := range sortedKeys(cfg.groups) {
		g := cfg.groups[key]
		st := e.states[key]
		if st == nil || !st.on {
			continue
		}

		patternFn, ok := patterns[st.pattern]
		if !ok {
			patternFn = patterns["off"]
		}

		interval := g.interval
		if interval <= 0 {
			interval = defaultInterval
		}
		elapsed := time.Since(st.startedAt)
		step := int(elapsed / interval)
		phase := float64(elapsed%interval) / float64(interval)

		count := len(g.fixtures)
		for i, fixtureKey := range g.fixtures {
			idx := i
			if g.reverse {
				idx = count - 1 - i
			}
			out := patternFn(frame{
				Index:   idx,
				Count:   count,
				Step:    step,
				Phase:   phase,
				Colors:  g.colors,
				Reverse: g.reverse,
			})

			fx, ok := cfg.fixtures[fixtureKey]
			if !ok {
				continue
			}
			writeFixture(buf, fx, out.Color, clamp01(out.Intensity*st.brightness))
		}
	}

	e.mu.Unlock()

	if err := e.output().Send(buf); err != nil {
		logrus.Warnf("dmx: send frame: %s", err)
		e.outputBroken.Store(true)
	}
}

// writeFixture writes one fixture's channels for the given color/intensity.
// If the fixture's profile has no dedicated dimmer channel, intensity is
// pre-multiplied into the color channels; otherwise the dimmer channel
// carries the level and the color channels are written at full value.
func writeFixture(buf []byte, fx resolvedFixture, color rgb, intensity float64) {
	start, _ := fx.span()

	colorScale := intensity
	if _, hasDimmer := fx.profile.roleOffset("dimmer"); hasDimmer {
		colorScale = 1
	}

	for offset, role := range fx.profile.channels {
		pos := start + offset
		if pos < 0 || pos >= len(buf) {
			continue
		}
		switch role {
		case "dimmer":
			buf[pos] = scaleByte(intensity)
		case "red":
			buf[pos] = scaleByte(color.R * colorScale)
		case "green":
			buf[pos] = scaleByte(color.G * colorScale)
		case "blue":
			buf[pos] = scaleByte(color.B * colorScale)
		case "white":
			white := math.Min(color.R, math.Min(color.G, color.B))
			buf[pos] = scaleByte(white * colorScale)
		}
		// "amber", "uv", "-" and any custom role are left as written by the
		// static pass above (or 0) — there is no color formula for them.
	}
}
