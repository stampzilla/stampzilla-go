package store

import "github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/models/devices"

func (store *Store) GetDevices() *devices.List {
	store.RLock()
	defer store.RUnlock()
	return store.Devices
}

func (store *Store) SyncState(list map[string]devices.State) {
	//TODO we should not set the state on the devices.
	//The node should get a state-change request and server will update devices state when node has done its job
	for id, state := range list {
		dev := store.Devices.GetUnique(id)
		if dev == nil {
			return
		}
		dev.SyncState(state)
	}
}
func (store *Store) AddOrUpdateDevice(dev *devices.Device) {
	store.Devices.Add(dev)
	store.Logic.UpdateDevice(dev)
	store.runCallbacks("devices")
}
