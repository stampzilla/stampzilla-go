package store

import "github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/models"

func (store *Store) GetDevices() *models.Devices {
	store.RLock()
	defer store.RUnlock()
	return store.Devices
}

func (store *Store) AddOrUpdateDevice(dev *models.Device) {
	store.Lock()
	store.Devices.Add(dev)
	store.Unlock()

	store.runCallbacks("devices")
}
