package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"
)

func TestExampleConfigIsValid(t *testing.T) {
	data, err := os.ReadFile("config.example.json")
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if _, err := loadConfig(data); err != nil {
		t.Fatalf("loadConfig(config.example.json) error = %v", err)
	}
}

func TestResolveConfigDefaults(t *testing.T) {
	data := []byte(`{
		"fixtures": {"f1": {"profile": "rgb", "address": 1}},
		"groups": {"g1": {"fixtures": ["f1"]}}
	}`)
	cfg, err := loadConfig(data)
	if err != nil {
		t.Fatalf("loadConfig() error = %v", err)
	}
	if cfg.fps != defaultFPS {
		t.Errorf("fps = %d, want %d", cfg.fps, defaultFPS)
	}
	g := cfg.groups["g1"]
	if g.pattern != "static" {
		t.Errorf("pattern = %q, want static", g.pattern)
	}
	if g.interval != defaultInterval {
		t.Errorf("interval = %v, want %v", g.interval, defaultInterval)
	}
	if len(g.colors) != 1 || g.colors[0] != (rgb{1, 1, 1}) {
		t.Errorf("colors = %v, want white default", g.colors)
	}
}

func TestResolveConfigBuiltinProfile(t *testing.T) {
	data := []byte(`{
		"fixtures": {"f1": {"profile": "rgbw-dimmer", "address": 10}},
		"groups": {"g1": {"fixtures": ["f1"]}}
	}`)
	cfg, err := loadConfig(data)
	if err != nil {
		t.Fatalf("loadConfig() error = %v", err)
	}
	fx := cfg.fixtures["f1"]
	if len(fx.profile.channels) != 5 {
		t.Fatalf("channels = %v, want 5", fx.profile.channels)
	}
	if _, ok := fx.profile.roleOffset("dimmer"); !ok {
		t.Errorf("expected dimmer role in rgbw-dimmer profile")
	}
}

func TestResolveConfigCustomProfileOverridesBuiltin(t *testing.T) {
	data := []byte(`{
		"profiles": {"rgb": {"channels": ["red", "green", "blue", "white"]}},
		"fixtures": {"f1": {"profile": "rgb", "address": 1}},
		"groups": {"g1": {"fixtures": ["f1"]}}
	}`)
	cfg, err := loadConfig(data)
	if err != nil {
		t.Fatalf("loadConfig() error = %v", err)
	}
	if len(cfg.fixtures["f1"].profile.channels) != 4 {
		t.Errorf("custom rgb profile was not applied, channels = %v", cfg.fixtures["f1"].profile.channels)
	}
}

func TestFPSClamped(t *testing.T) {
	data := []byte(`{"fps": 1000, "fixtures": {}, "groups": {}}`)
	cfg, err := loadConfig(data)
	if err != nil {
		t.Fatalf("loadConfig() error = %v", err)
	}
	if cfg.fps != maxFPS {
		t.Errorf("fps = %d, want clamped to %d", cfg.fps, maxFPS)
	}
}

func TestResolveConfigBreakModeAndDEModeDefaults(t *testing.T) {
	data := []byte(`{"fixtures": {}, "groups": {}}`)
	cfg, err := loadConfig(data)
	if err != nil {
		t.Fatalf("loadConfig() error = %v", err)
	}
	if cfg.breakMode != breakModeBaud {
		t.Errorf("breakMode = %q, want default %q", cfg.breakMode, breakModeBaud)
	}
	if cfg.deMode != deModeNone {
		t.Errorf("deMode = %q, want default %q", cfg.deMode, deModeNone)
	}
}

func TestResolveConfigBreakModeAndDEModeExplicit(t *testing.T) {
	data := []byte(`{"breakMode": "ioctl", "deMode": "assert", "fixtures": {}, "groups": {}}`)
	cfg, err := loadConfig(data)
	if err != nil {
		t.Fatalf("loadConfig() error = %v", err)
	}
	if cfg.breakMode != breakModeIoctl {
		t.Errorf("breakMode = %q, want %q", cfg.breakMode, breakModeIoctl)
	}
	if cfg.deMode != deModeAssert {
		t.Errorf("deMode = %q, want %q", cfg.deMode, deModeAssert)
	}
}

func TestResolveConfigUniverseSizeExplicit(t *testing.T) {
	data := []byte(`{
		"universeSize": 512,
		"fixtures": {"f1": {"profile": "rgb", "address": 1}},
		"groups": {"g1": {"fixtures": ["f1"]}}
	}`)
	cfg, err := loadConfig(data)
	if err != nil {
		t.Fatalf("loadConfig() error = %v", err)
	}
	if cfg.universeSize != dmxUniverseChannels {
		t.Errorf("universeSize = %d, want %d", cfg.universeSize, dmxUniverseChannels)
	}
}

func TestResolveConfigFPSClampedForLargeUniverse(t *testing.T) {
	data := []byte(`{"fps": 44, "universeSize": 512, "fixtures": {}, "groups": {}}`)
	cfg, err := loadConfig(data)
	if err != nil {
		t.Fatalf("loadConfig() error = %v", err)
	}
	if cfg.fps >= 44 {
		t.Errorf("fps = %d, want clamped below 44 for a 512-channel universe", cfg.fps)
	}
}

func TestResolveConfigErrors(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr string
	}{
		{
			"unknown profile",
			`{"fixtures":{"f1":{"profile":"nope","address":1}},"groups":{}}`,
			"unknown profile",
		},
		{
			"unknown pattern",
			`{"fixtures":{"f1":{"profile":"rgb","address":1}},"groups":{"g1":{"fixtures":["f1"],"pattern":"nope"}}}`,
			"unknown pattern",
		},
		{
			"address too low",
			`{"fixtures":{"f1":{"profile":"rgb","address":0}},"groups":{}}`,
			"address must be",
		},
		{
			"address overflows universe",
			`{"fixtures":{"f1":{"profile":"rgb","address":511}},"groups":{}}`,
			"overflows",
		},
		{
			"overlapping fixtures",
			`{"fixtures":{"f1":{"profile":"rgb","address":1},"f2":{"profile":"rgb","address":2}},"groups":{}}`,
			"overlap",
		},
		{
			"group references unknown fixture",
			`{"fixtures":{},"groups":{"g1":{"fixtures":["nope"]}}}`,
			"unknown fixture",
		},
		{
			"empty group",
			`{"fixtures":{},"groups":{"g1":{"fixtures":[]}}}`,
			"at least one fixture",
		},
		{
			"bad color",
			`{"fixtures":{"f1":{"profile":"rgb","address":1}},"groups":{"g1":{"fixtures":["f1"],"colors":["not-a-color"]}}}`,
			"invalid color",
		},
		{
			"bad interval",
			`{"fixtures":{"f1":{"profile":"rgb","address":1}},"groups":{"g1":{"fixtures":["f1"],"interval":"not-a-duration"}}}`,
			"invalid duration",
		},
		{
			"profile with no channels",
			`{"profiles":{"empty":{"channels":[]}},"fixtures":{"f1":{"profile":"empty","address":1}},"groups":{}}`,
			"at least one channel",
		},
		{
			"static references undeclared channel",
			`{"profiles":{"p":{"channels":["dimmer"],"static":{"mode":1}}},"fixtures":{"f1":{"profile":"p","address":1}},"groups":{}}`,
			"is not declared",
		},
		{
			"unknown breakMode",
			`{"breakMode":"nope","fixtures":{},"groups":{}}`,
			"unknown breakMode",
		},
		{
			"unknown deMode",
			`{"deMode":"nope","fixtures":{},"groups":{}}`,
			"unknown deMode",
		},
		{
			"universeSize too large",
			`{"universeSize":600,"fixtures":{},"groups":{}}`,
			"universeSize must be",
		},
		{
			"universeSize smaller than fixtures need",
			`{"universeSize":5,"fixtures":{"f1":{"profile":"rgb","address":1}},"groups":{}}`,
			"is smaller than",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := loadConfig([]byte(tt.json))
			if err == nil {
				t.Fatalf("loadConfig() error = nil, want containing %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("loadConfig() error = %q, want containing %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestDurationUnmarshal(t *testing.T) {
	var d duration
	if err := json.Unmarshal([]byte(`"400ms"`), &d); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if time.Duration(d) != 400*time.Millisecond {
		t.Errorf("duration = %v, want 400ms", time.Duration(d))
	}
}

func TestDurationUnmarshalEmpty(t *testing.T) {
	var d duration
	if err := json.Unmarshal([]byte(`""`), &d); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if d != 0 {
		t.Errorf("duration = %v, want 0", time.Duration(d))
	}
}
