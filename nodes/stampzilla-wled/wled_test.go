package main

import (
	"testing"
)

func TestBriFromFloat(t *testing.T) {
	tests := []struct {
		b    float64
		want int
	}{
		{0.0, 0},
		{1.0, 255},
		{0.5, 128}, // math.Round(0.5*255) = math.Round(127.5) = 128
		{0.25, 64}, // math.Round(63.75) = 64
		{-0.5, 0},  // clamped to 0
		{1.5, 255}, // clamped to 1 → 255
	}
	for _, tt := range tests {
		got := briFromFloat(tt.b)
		if got != tt.want {
			t.Errorf("briFromFloat(%v) = %d, want %d", tt.b, got, tt.want)
		}
	}
}

func TestBriDecode(t *testing.T) {
	tests := []struct {
		bri  int
		want float64
	}{
		{0, 0.0},
		{255, 1.0},
		{128, 128.0 / 255.0},
		{64, 64.0 / 255.0},
	}
	for _, tt := range tests {
		got := float64(tt.bri) / 255.0
		if got != tt.want {
			t.Errorf("bri decode(%d) = %v, want %v", tt.bri, got, tt.want)
		}
	}
}

func TestWsURL(t *testing.T) {
	tests := []struct {
		host string
		want string
	}{
		{"192.168.1.100", "ws://192.168.1.100/ws"},
		{"http://192.168.1.100", "ws://192.168.1.100/ws"},
		{"https://192.168.1.100", "wss://192.168.1.100/ws"},
		{"ws://192.168.1.100", "ws://192.168.1.100/ws"},
		{"wss://192.168.1.100", "wss://192.168.1.100/ws"},
		{"ws://192.168.1.100/ws", "ws://192.168.1.100/ws"},
		{"http://192.168.1.100/", "ws://192.168.1.100/ws"},
	}
	for _, tt := range tests {
		got := wsURL(tt.host)
		if got != tt.want {
			t.Errorf("wsURL(%q) = %q, want %q", tt.host, got, tt.want)
		}
	}
}

func TestParseWLEDState(t *testing.T) {
	msg := []byte(`{"state":{"on":true,"bri":127},"info":{"ver":"0.14.0"}}`)
	state, err := parseWLEDState(msg)
	if err != nil {
		t.Fatalf("parseWLEDState: unexpected error: %s", err)
	}
	if on, ok := state["on"].(bool); !ok || !on {
		t.Errorf("state[on] = %v, want true", state["on"])
	}
	wantBrightness := 127.0 / 255.0
	if b, ok := state["brightness"].(float64); !ok || b != wantBrightness {
		t.Errorf("state[brightness] = %v, want %v", state["brightness"], wantBrightness)
	}

	// off, bri=0
	msg2 := []byte(`{"state":{"on":false,"bri":0}}`)
	state2, err := parseWLEDState(msg2)
	if err != nil {
		t.Fatalf("parseWLEDState: %s", err)
	}
	if on, ok := state2["on"].(bool); !ok || on {
		t.Errorf("state2[on] = %v, want false", state2["on"])
	}
	if b, ok := state2["brightness"].(float64); !ok || b != 0.0 {
		t.Errorf("state2[brightness] = %v, want 0.0", state2["brightness"])
	}

	// invalid JSON
	_, err = parseWLEDState([]byte(`not json`))
	if err == nil {
		t.Error("parseWLEDState(invalid): expected error, got nil")
	}
}
