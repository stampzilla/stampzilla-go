package main

import (
	"testing"
)

func TestVolumeToMV(t *testing.T) {
	tests := []struct {
		v    float64
		want string
	}{
		{0.0, "MV00"},
		{1.0, "MV80"},
		{0.5, "MV40"},
		{0.38125, "MV305"}, // 80 * 0.38125 = 30.5 → half step
		{0.75625, "MV605"}, // 80 * 0.75625 = 60.5 → half step
		{-0.1, "MV00"},     // clamped below 0
		{1.5, "MV80"},      // clamped above 1
	}
	for _, tt := range tests {
		got := volumeToMV(tt.v)
		if got != tt.want {
			t.Errorf("volumeToMV(%v) = %q, want %q", tt.v, got, tt.want)
		}
	}
}

func TestMVToVolume(t *testing.T) {
	tests := []struct {
		param string
		want  float64
	}{
		{"00", 0.0},
		{"80", 1.0},
		{"40", 0.5},
		{"805", 1.0},     // 80.5/80 clamped to 1.0
		{"395", 0.49375}, // 39.5/80
		{"305", 0.38125}, // 30.5/80
	}
	for _, tt := range tests {
		got := mvToVolume(tt.param)
		if got != tt.want {
			t.Errorf("mvToVolume(%q) = %v, want %v", tt.param, got, tt.want)
		}
	}

	// invalid inputs
	for _, bad := range []string{"", "x0", "abc", "8", "1234"} {
		if v := mvToVolume(bad); v != -1 {
			t.Errorf("mvToVolume(%q) = %v, want -1", bad, v)
		}
	}
}

func TestParseEvent(t *testing.T) {
	tests := []struct {
		line  string
		key   string
		value any
	}{
		{"PWON", "on", true},
		{"PWSTANDBY", "on", false},
		{"MV80", "volume", 1.0},
		{"MV00", "volume", 0.0},
		{"MV40", "volume", 0.5},
		{"MV805", "volume", 1.0}, // 80.5/80 clamped
		{"SIDVD", "source", "DVD"},
		{"SIMPLAY", "source", "MPLAY"},
		{"SITV", "source", "TV"},
		{"SICD", "source", "CD"},
	}
	for _, tt := range tests {
		state := parseEvent(tt.line)
		if state == nil {
			t.Errorf("parseEvent(%q) returned nil, want state with %q=%v", tt.line, tt.key, tt.value)
			continue
		}
		got, ok := state[tt.key]
		if !ok {
			t.Errorf("parseEvent(%q): missing key %q", tt.line, tt.key)
			continue
		}
		if got != tt.value {
			t.Errorf("parseEvent(%q)[%q] = %v (%T), want %v (%T)", tt.line, tt.key, got, got, tt.value, tt.value)
		}
	}

	// These lines should produce no state.
	ignored := []string{
		"MVMAX 98",
		"MVMAX98",
		"MS STEREO",
		"DC AUTO",
		"",
		"PWMUTING",
	}
	for _, line := range ignored {
		if state := parseEvent(line); len(state) != 0 {
			t.Errorf("parseEvent(%q) should return nil/empty, got %v", line, state)
		}
	}
}
