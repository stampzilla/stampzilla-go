package protocol

import (
	"sync"

	"github.com/stampzilla/stampzilla-go/protocol/devices"
)

type Devices struct {
	devices devices.Map
	sync.RWMutex
}

func NewDevices() *Devices {
	n := &Devices{}
	n.devices = make(devices.Map)
	return n
}

func (n *Devices) ByUuid(uuid string) *devices.Device {
	n.RLock()
	defer n.RUnlock()
	if node, ok := n.devices[uuid]; ok {
		return node
	}
	return nil
}
func (n *Devices) All() map[string]devices.Device {
	n.RLock()
	defer n.RUnlock()
	r := make(map[string]devices.Device)
	for k, v := range n.devices {
		r[k] = *v
	}
	return r
}
func (n *Devices) AllWithState(nodes *Nodes) map[string]devices.Device {
	devices := n.All()
	for _, device := range devices {
		device.SyncState(nodes.ByUuid(device.Node).State())
	}
	return devices
}
func (n *Devices) Add(nodeUuid string, device *devices.Device) error {
	n.Lock()
	defer n.Unlock()
	n.devices[nodeUuid+"."+device.Id] = device
	return nil
}
func (n *Devices) Delete(uuid string) {
	n.Lock()
	defer n.Unlock()
	delete(n.devices, uuid)
}
func (n *Devices) SetOfflineByNode(node string) Node {
	n.RLock()
	defer n.RUnlock()
	for _, dev := range n.devices {
		if dev.Node == node {
			dev.Online = false
		}
	}
	return nil
}
