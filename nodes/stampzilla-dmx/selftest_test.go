package main

import (
	"bytes"
	"flag"
	"strings"
	"testing"
	"time"
)

// TestSelftestRequestedDoesNotShadowNodeFlags is the regression test for the
// original bug report: a normal invocation (any real node flag, --help,
// -version, or no args at all) must never be mistaken for the selftest
// subcommand, or the node's own flags (parsed later by multiconfig inside
// node.New()) never get a chance to run.
func TestSelftestRequestedDoesNotShadowNodeFlags(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{"no args", nil, false},
		{"host flag", []string{"-host", "example.com:8080"}, false},
		{"help", []string{"--help"}, false},
		{"version", []string{"-version"}, false},
		{"old top-level selftest flag", []string{"-selftest-port", "/dev/ttyUSB0"}, false},
		{"selftest subcommand", []string{"selftest", "-port", "/dev/ttyUSB0"}, true},
		{"selftest subcommand alone", []string{"selftest"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := selftestRequested(tt.args); got != tt.want {
				t.Errorf("selftestRequested(%v) = %v, want %v", tt.args, got, tt.want)
			}
		})
	}
}

func TestLegacySelftestHint(t *testing.T) {
	if got := legacySelftestHint([]string{"-selftest-port", "/dev/ttyUSB0"}); got == "" {
		t.Error("legacySelftestHint() = \"\", want a hint for the old -selftest-* syntax")
	}
	if got := legacySelftestHint([]string{"--selftest-mode", "full"}); got == "" {
		t.Error("legacySelftestHint() = \"\", want a hint for the old --selftest-* syntax")
	}
	if got := legacySelftestHint([]string{"-host", "example.com"}); got != "" {
		t.Errorf("legacySelftestHint() = %q, want \"\" for normal flags", got)
	}
	if got := legacySelftestHint(nil); got != "" {
		t.Errorf("legacySelftestHint() = %q, want \"\" for no args", got)
	}
}

func TestParseSelftestArgsDefaults(t *testing.T) {
	var out bytes.Buffer
	cfg, err := parseSelftestArgs([]string{"-port", "/dev/ttyUSB0"}, &out)
	if err != nil {
		t.Fatalf("parseSelftestArgs() error = %v", err)
	}
	if cfg.port != "/dev/ttyUSB0" {
		t.Errorf("port = %q, want /dev/ttyUSB0", cfg.port)
	}
	if cfg.mode != "full" {
		t.Errorf("mode = %q, want full", cfg.mode)
	}
	if cfg.channels != minChannels {
		t.Errorf("channels = %d, want %d", cfg.channels, minChannels)
	}
	if cfg.fps != defaultFPS {
		t.Errorf("fps = %d, want %d", cfg.fps, defaultFPS)
	}
	if cfg.breakMode != breakModeBaud {
		t.Errorf("breakMode = %q, want %q", cfg.breakMode, breakModeBaud)
	}
	if cfg.deMode != deModeNone {
		t.Errorf("deMode = %q, want %q", cfg.deMode, deModeNone)
	}
	if cfg.echo {
		t.Error("echo = true, want false by default")
	}
	if cfg.walkHold != 2*time.Second {
		t.Errorf("walkHold = %v, want 2s", cfg.walkHold)
	}
}

func TestParseSelftestArgsAllFlags(t *testing.T) {
	var out bytes.Buffer
	args := []string{
		"-port", "/dev/ttyUSB1",
		"-mode", "walk",
		"-channels", "10",
		"-fps", "20",
		"-break-mode", "ioctl",
		"-de-mode", "assert",
		"-echo",
		"-walk-hold", "500ms",
	}
	cfg, err := parseSelftestArgs(args, &out)
	if err != nil {
		t.Fatalf("parseSelftestArgs() error = %v", err)
	}
	if cfg.port != "/dev/ttyUSB1" || cfg.mode != "walk" || cfg.channels != 10 || cfg.fps != 20 {
		t.Errorf("cfg = %+v, unexpected", cfg)
	}
	if cfg.breakMode != breakModeIoctl {
		t.Errorf("breakMode = %q, want %q", cfg.breakMode, breakModeIoctl)
	}
	if cfg.deMode != deModeAssert {
		t.Errorf("deMode = %q, want %q", cfg.deMode, deModeAssert)
	}
	if !cfg.echo {
		t.Error("echo = false, want true")
	}
	if cfg.walkHold != 500*time.Millisecond {
		t.Errorf("walkHold = %v, want 500ms", cfg.walkHold)
	}
}

func TestParseSelftestArgsHelp(t *testing.T) {
	var out bytes.Buffer
	_, err := parseSelftestArgs([]string{"-h"}, &out)
	if err != flag.ErrHelp {
		t.Errorf("parseSelftestArgs(-h) error = %v, want flag.ErrHelp", err)
	}
	if out.Len() == 0 {
		t.Error("parseSelftestArgs(-h) printed no usage")
	}
}

func TestParseSelftestArgsErrors(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{"missing port", nil, "-port is required"},
		{"unknown flag", []string{"-bogus"}, "flag provided but not defined"},
		{"bad channels", []string{"-port", "/dev/ttyUSB0", "-channels", "0"}, "-channels must be"},
		{"bad fps", []string{"-port", "/dev/ttyUSB0", "-fps", "0"}, "-fps must be"},
		{"bad mode", []string{"-port", "/dev/ttyUSB0", "-mode", "nope"}, "unknown -mode"},
		{"bad break mode", []string{"-port", "/dev/ttyUSB0", "-break-mode", "nope"}, "unknown -break-mode"},
		{"bad de mode", []string{"-port", "/dev/ttyUSB0", "-de-mode", "nope"}, "unknown -de-mode"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			_, err := parseSelftestArgs(tt.args, &out)
			if err == nil {
				t.Fatalf("parseSelftestArgs() error = nil, want containing %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("parseSelftestArgs() error = %q, want containing %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestSelftestFrameFull(t *testing.T) {
	cfg := &selftestConfig{mode: "full", channels: 5}
	frame := selftestFrame(cfg, 0)
	if len(frame) != 5 {
		t.Fatalf("len(frame) = %d, want 5", len(frame))
	}
	for i, v := range frame {
		if v != 255 {
			t.Errorf("frame[%d] = %d, want 255", i, v)
		}
	}
}

func TestSelftestFrameWalkLightsExactlyOneChannel(t *testing.T) {
	cfg := &selftestConfig{mode: "walk", channels: 4, walkHold: time.Second}

	tests := []struct {
		elapsed time.Duration
		want    int
	}{
		{0, 0},
		{999 * time.Millisecond, 0},
		{1 * time.Second, 1},
		{2*time.Second + 500*time.Millisecond, 2},
		{4 * time.Second, 0}, // wraps back around after 4 channels
	}
	for _, tt := range tests {
		frame := selftestFrame(cfg, tt.elapsed)
		if len(frame) != 4 {
			t.Fatalf("len(frame) = %d, want 4", len(frame))
		}
		nonZero := 0
		for i, v := range frame {
			if v != 0 {
				nonZero++
				if i != tt.want {
					t.Errorf("elapsed=%v: channel %d is on, want channel %d", tt.elapsed, i, tt.want)
				}
				if v != 255 {
					t.Errorf("elapsed=%v: channel %d = %d, want 255", tt.elapsed, i, v)
				}
			}
		}
		if nonZero != 1 {
			t.Errorf("elapsed=%v: %d channels non-zero, want exactly 1", tt.elapsed, nonZero)
		}
	}
}

func TestSelftestFrameRampStaysInRange(t *testing.T) {
	cfg := &selftestConfig{mode: "ramp", channels: 3}
	for _, elapsed := range []time.Duration{0, 500 * time.Millisecond, 1 * time.Second, 1500 * time.Millisecond, 3 * time.Second} {
		frame := selftestFrame(cfg, elapsed)
		v := frame[0]
		for _, other := range frame {
			if other != v {
				t.Fatalf("elapsed=%v: ramp channels differ: %v", elapsed, frame)
			}
		}
		// Byte range is implicit ([0,255]), but check the midpoint of the
		// period is near full brightness and the boundaries are near zero.
	}
	cfg2 := &selftestConfig{mode: "ramp", channels: 1}
	mid := selftestFrame(cfg2, 1*time.Second)[0] // half of the 2s period
	if mid < 200 {
		t.Errorf("ramp at period midpoint = %d, want near 255", mid)
	}
	start := selftestFrame(cfg2, 0)[0]
	if start != 0 {
		t.Errorf("ramp at period start = %d, want 0", start)
	}
}

func TestWalkChannelDefaultsHoldIfZero(t *testing.T) {
	cfg := &selftestConfig{channels: 3, walkHold: 0}
	// Should not panic (divide by zero) and should fall back to a sane default.
	_ = walkChannel(cfg, 3*time.Second)
}
