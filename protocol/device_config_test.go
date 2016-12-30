package protocol

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockIdentifiable struct {
	uuid string
}

func (i mockIdentifiable) Uuid() string {
	return i.uuid
}

func TestConfigMapHandlerIsCalled(t *testing.T) {

	mockIdentifiable := mockIdentifiable{
		uuid: "asdf",
	}
	dcm := NewConfigMap(mockIdentifiable)

	handlerRan := false
	dcm.Add("device1").Layout(
		&DeviceConfig{
			ID:   "46",
			Name: "LOAD ERROR Alarmreport",
			Options: map[string]string{
				"0": "No reaction",
				"1": "Send an alarm frame",
			},
		},
		&DeviceConfig{
			ID:   "47",
			Name: "Ignorera",
			Type: "bool",
		},
		&DeviceConfig{
			ID:   "48",
			Name: "Ignorera",
			Type: "float",
			Min:  0,
			Max:  99,
		},
	).Handler(func(device string, c *DeviceConfig) {
		assert.Equal(t, "device1", device)
		assert.Equal(t, "47", c.ID)
		assert.Equal(t, 123, c.Value)
		handlerRan = true
	})

	assert.Equal(t, false, handlerRan)

	dcm.SetConfig("device1", "47", 123)

	assert.Equal(t, true, handlerRan)
}
