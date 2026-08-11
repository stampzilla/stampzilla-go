package main

import "testing"

func TestClampChannels(t *testing.T) {
	tests := []struct {
		name     string
		channels []byte
		wantLen  int
	}{
		{"nil channels padded to floor", nil, minChannels},
		{"below floor padded", []byte{1, 2, 3}, minChannels},
		{"exact floor", make([]byte, minChannels), minChannels},
		{"above floor kept as-is", make([]byte, 100), 100},
		{"clamped to max", make([]byte, 600), maxChannels},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := clampChannels(tt.channels)
			if len(got) != tt.wantLen {
				t.Fatalf("len(clampChannels(%d channels)) = %d, want %d", len(tt.channels), len(got), tt.wantLen)
			}
		})
	}
}

func TestClampChannelsPreservesValues(t *testing.T) {
	in := []byte{1, 2, 3}
	got := clampChannels(in)
	for i, v := range in {
		if got[i] != v {
			t.Errorf("clampChannels()[%d] = %d, want %d", i, got[i], v)
		}
	}
}

func TestLogOutput(t *testing.T) {
	var o dmxOutput = logOutput{}
	if err := o.Send([]byte{1, 2, 3}); err != nil {
		t.Errorf("Send() error = %v", err)
	}
	if err := o.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}
}
