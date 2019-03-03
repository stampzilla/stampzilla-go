package store

import "github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models/devices"

func (store *Store) GetDevices() *devices.List {
	store.RLock()
	defer store.RUnlock()
	return store.Devices
}

func (store *Store) AddOrUpdateDevice(dev *devices.Device) {
	if dev == nil {
		return
	}

	oldDev := store.Devices.Get(dev.ID)
	if oldDev != nil && oldDev.Equal(dev) {
		return
	}

	store.Devices.Add(dev)
	node := store.GetNode(dev.ID.Node)

	alias := node.Alias(dev.ID)
	if alias != dev.Alias {
		dev.Lock()
		dev.Alias = alias
		dev.Unlock()
	}

	store.Logic.UpdateDevice(dev)
	store.runCallbacks("devices")
}
