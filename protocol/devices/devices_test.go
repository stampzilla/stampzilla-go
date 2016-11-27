package devices

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSyncState(t *testing.T) {

	device := NewDevice()
	device.Name = "Test"
	device.StateMap["On"] = "Devices[1].On"
	device.StateMap["Off"] = "Devices[2].On"
	state := `
		{
			"Devices": {
				"1": {
					"Id": "1",
					"Name": "Dev1",
					"On": true,
					"Type": ""
				},
				"2": {
					"Id": "2",
					"Name": "Dev2",
					"On": false,
					"Type": ""
				}
			}
		}
	`
	var v interface{}
	err := json.Unmarshal([]byte(state), &v)
	assert.NoError(t, err)

	device.SyncState(v)
	assert.Equal(t, true, device.State["On"])
	assert.Equal(t, false, device.State["Off"])
}
