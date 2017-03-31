package protocol

import (
	"testing"

	"github.com/stampzilla/stampzilla-go/protocol/devices"
	"github.com/stretchr/testify/assert"
)

func TestAdd(t *testing.T) {
	devs := NewDevices()
	device := devices.NewDevice()
	device.Id = "devuuid"
	device.Node = "nodeuuid"
	devs.Add(device)

	if dev := devs.devices.ByID("nodeuuid.devuuid"); dev != nil {
		assert.Equal(t, "devuuid", dev.Id)
		return
	}

	t.Fatalf("Device nodeuuid.devuuid was not found")
}

func TestShallowCopy(t *testing.T) {
	devs := NewDevices()
	device := devices.NewDevice()
	device.Id = "devuuid"
	device.Name = "Name"
	device.Node = "nodeuuid"
	devs.Add(device)

	copied := devs.ShallowCopy()
	if _, ok := copied["nodeuuid.devuuid"]; !ok {
		t.Fatal("nodeuuid.devuuid not found")
	}
	device.Name = "NameChange"
	assert.Equal(t, "Name", copied["nodeuuid.devuuid"].Name)
	dev := devs.devices.ByID("nodeuuid.devuuid")
	if assert.NotNil(t, dev) {
		assert.Equal(t, "NameChange", dev.Name)
	}
}

func TestSetOfflineByNode(t *testing.T) {
	devs := NewDevices()

	device1 := devices.NewDevice()
	device1.Id = "1"
	device1.Node = "node1"
	device1.Online = true

	device2 := devices.NewDevice()
	device2.Id = "2"
	device2.Node = "node2"
	device2.Online = true

	devs.Add(device1)
	devs.Add(device2)

	list := devs.SetOfflineByNode("node1")

	assert.Len(t, list, 1)
	assert.Equal(t, false, device1.Online)
}
