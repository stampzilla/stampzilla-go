package main

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHeatPumpState(t *testing.T) {
	tests := []struct {
		in  string
		out string
	}{
		// 65529 - 65536 = -7
		{`{"0007":65529}`, `"Outdoor":-0.7`},
		{`{"0007":123}`, `"Outdoor":12.3`},
		{`{"1A01":1}`, `"Compressor":true`},
		{`{"1A01":0}`, `"Compressor":false`},
	}

	for _, tt := range tests {
		data := tt
		t.Run(data.in, func(t *testing.T) {
			hp := &HeatPump{}
			err := json.Unmarshal([]byte(data.in), hp)
			assert.NoError(t, err)
			out, err := json.Marshal(hp.State())
			assert.NoError(t, err)
			assert.Contains(t, string(out), data.out)
		})
	}
}
