package store

import "github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models/devices"

func (store *Store) GetDevices() *devices.List {
	store.RLock()
	defer store.RUnlock()
	return store.Devices
}

func (store *Store) AddOrUpdateDevice(dev *devices.Device) {
	store.Devices.Add(dev)
	store.Logic.UpdateDevice(dev)
	store.runCallbacks("devices")
}
