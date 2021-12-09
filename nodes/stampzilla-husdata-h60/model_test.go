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
		{`{"0007":-7}`, `"Outdoor":-0.7`},
		{`{"0007":-71}`, `"Outdoor":-7.1`},
		{`{"0007":123}`, `"Outdoor":12.3`},
		{`{"1A01":1}`, `"Compressor":true`},
		{`{"1A01":0}`, `"Compressor":false`},
		{`{"5C52":844329}`, `"SuppliedEnergyHeating":8443.29`},
		{`{"5C53":463482}`, `"SuppliedEnergyHotwater":4634.82`},
		{`{"5C55":4758500}`, `"CompressorConsumptionHeating":4758.5`},
		{`{"5C56":2736630}`, `"CompressorConsumptionHotwater":2736.63`},
		{`{"5C58":3608}`, `"AuxConsumptionHeating":3.608`},
		{`{"5C59":250314}`, `"AuxConsumptionHotwater":250.314`},
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
