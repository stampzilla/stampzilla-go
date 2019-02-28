package store

import "github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models/devices"

func (store *Store) GetDevices() *devices.List {
	store.RLock()
	defer store.RUnlock()
	return store.Devices
}

func (store *Store) AddOrUpdateDevice(dev *devices.Device) {
	store.Devices.Add(dev)
	node := store.GetNode(dev.ID.Node)

	node.Lock()
	if a, ok := node.Aliases[dev.ID]; ok {
		dev.Lock()
		dev.Alias = a
		dev.Unlock()
	}
	node.Unlock()

	store.Logic.UpdateDevice(dev)
	store.runCallbacks("devices")
}
