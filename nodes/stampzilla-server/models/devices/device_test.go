package devices

import (
	"encoding/json"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func testDevice(id string) *Device {
	return &Device{
		Type: "type",
		//Node:   "node",
		ID: ID{
			ID:   id,
			Node: "node",
		},
		Name:   "name",
		Online: true,
		State: State{
			"on": true,
		},
		Traits: []string{"onoff"},
	}
}

func TestCopyDevice(t *testing.T) {
	d := testDevice("id")

	newD := d.Copy()

	d.Type = "0"
	//d.Node = "0"
	d.ID = ID{
		ID:   "0",
		Node: "0",
	}
	d.Name = "0"
	d.Online = false
	d.State["on"] = false
	d.State["off"] = true
	d.Traits = append(d.Traits, "0")

	assert.Equal(t, "type", newD.Type)
	assert.Equal(t, "0", d.Type)
	assert.Equal(t, "node", newD.ID.Node)
	assert.Equal(t, "id", newD.ID.ID)
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
	devices := NewList()
	devices.Add(d)
	newD := devices.Copy()
	devices.Add(testDevice("id2"))

	assert.Len(t, newD.devices, 1)
	assert.Len(t, devices.devices, 2)
}

func TestFlatten(t *testing.T) {
	d := testDevice("id")
	d.State["temperature"] = 10
	devices := NewList()
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
	ds1 := State{
		"on":          true,
		"temperature": 10,
		"test":        1,
	}
	ds2 := State{
		"on":          false,
		"temperature": 10,
		"asdf":        10,
	}
	diff := ds1.Diff(ds2)
	assert.Len(t, diff, 2)
	//spew.Dump(diff)
	assert.Equal(t, false, diff["on"])
	assert.Equal(t, 10, diff["asdf"])
	assert.Equal(t, nil, diff["temperature"])
}

func TestJSONMarshalDevices(t *testing.T) {
	d := NewList()

	d.Add(testDevice("devid"))

	b, err := json.MarshalIndent(d, "", "\t")
	if err != nil {
		logrus.Error("error marshal json", err)
	}

	assert.Contains(t, string(b), `"node.devid": {`)
	assert.NoError(t, err)
}
func TestJSONUnMarshalDevices(t *testing.T) {
	j := `{
	"node.devid": {
		"type": "type",
		"id": "node.devid",
		"name": "name",
		"online": true,
		"state": {
			"on": true
		},
		"traits": [
			"onoff"
		]
	}}`
	d := NewList()

	err := json.Unmarshal([]byte(j), d)

	assert.NoError(t, err)
	assert.Equal(t, "devid", d.Get(ID{ID: "devid", Node: "node"}).ID.ID)
	assert.Equal(t, "node", d.Get(ID{ID: "devid", Node: "node"}).ID.Node)
	assert.Len(t, d.All(), 1)
}

func TestNewIDFromString(t *testing.T) {
	id, err := NewIDFromString("asdf.1")
	assert.NoError(t, err)
	assert.Equal(t, "1", id.ID)
	assert.Equal(t, "asdf", id.Node)
}

func TestDeviceEqual(t *testing.T) {
	dev1 := testDevice("1")
	dev2 := testDevice("1")
	eq := dev1.Equal(dev2)
	assert.Equal(t, true, eq)

	dev1 = testDevice("1")
	dev2 = testDevice("2")
	eq = dev1.Equal(dev2)
	assert.Equal(t, true, eq)

	dev1 = testDevice("1")
	dev2 = testDevice("2")
	dev2.Alias = "test"
	eq = dev1.Equal(dev2)
	assert.Equal(t, false, eq)

	dev1 = testDevice("1")
	dev2 = testDevice("2")
	dev2.State["on"] = false
	eq = dev1.Equal(dev2)
	assert.Equal(t, false, eq)

	dev1 = testDevice("1")
	dev2 = testDevice("2")
	dev2.State["newkey"] = false
	eq = dev1.Equal(dev2)
	assert.Equal(t, false, eq)
}
