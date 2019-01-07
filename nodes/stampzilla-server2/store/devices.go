package store

import "github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/models"

func (store *Store) GetDevices() *models.Devices {
	store.RLock()
	defer store.RUnlock()
	return store.Devices
}

func (store *Store) SyncState(list map[string]models.DeviceState) {
	for id, state := range list {
		dev := store.Devices.GetUnique(id)
		if dev == nil {
			return
		}
		dev.SyncState(state)
	}
	store.runCallbacks("devices")
}
func (store *Store) AddOrUpdateDevice(dev *models.Device) {
	store.Devices.Add(dev)
	store.runCallbacks("devices")
}
