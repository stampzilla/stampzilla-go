package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func testDevice(id string) *Device {
	return &Device{
		Type:   "type",
		Node:   "node",
		ID:     id,
		Name:   "name",
		Online: true,
		State: DeviceState{
			"on": true,
		},
		Traits: []string{"onoff"},
	}
}

func TestCopyDevice(t *testing.T) {
	d := testDevice("id")

	newD := d.Copy()

	d.Type = "0"
	d.Node = "0"
	d.ID = "0"
	d.Name = "0"
	d.Online = false
	d.State["on"] = false
	d.State["off"] = true
	d.Traits = append(d.Traits, "0")

	assert.Equal(t, "type", newD.Type)
	assert.Equal(t, "0", d.Type)
	assert.Equal(t, "node", newD.Node)
	assert.Equal(t, "id", newD.ID)
	assert.Equal(t, "name", newD.Name)
	assert.Equal(t, true, newD.Online)
	assert.Equal(t, true, newD.State["on"])

	assert.Len(t, newD.Traits, 1)
	assert.Len(t, newD.State, 1)
	assert.Len(t, d.Traits, 2)
	assert.Len(t, d.State, 2)

}

func TestCopyDevices(t *testing.T) {
	d := testDevice("id")
	devices := NewDevices()
	devices.Add(d)
	newD := devices.Copy()
	devices.Add(testDevice("id2"))

	assert.Len(t, newD.devices, 1)
	assert.Len(t, devices.devices, 2)
}

func TestFlatten(t *testing.T) {
	d := testDevice("id")
	d.State["temperature"] = 10
	devices := NewDevices()
	devices.Add(d)
	f := devices.Flatten()
	t.Log(f)
	tests := []struct {
		key      string
		expected interface{}
	}{
		{key: "node.id.on", expected: true},
		{key: "node.id.temperature", expected: 10},
	}
	for _, v := range tests {
		assert.Contains(t, f, v.key)
		assert.Equal(t, v.expected, f[v.key])
	}
}

func TestStateDiff(t *testing.T) {
	ds1 := DeviceState{
		"on":          true,
		"temperature": 10,
	}
	ds2 := DeviceState{
		"on":          false,
		"temperature": 10,
	}
	diff := ds1.Diff(ds2)
	assert.Equal(t, false, diff["on"])
	assert.Equal(t, nil, diff["temperature"])
}
