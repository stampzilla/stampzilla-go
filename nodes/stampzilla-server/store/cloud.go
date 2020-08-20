package store

import (
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models"
)

func (store *Store) GetCloud() models.Cloud {
	store.RLock()
	defer store.RUnlock()
	return store.Cloud
}

func (store *Store) UpdateCloudConfig(config models.CloudConfig) {
	store.RLock()
	if store.Cloud.Config.Equal(config) {
		store.RUnlock()
		return
	}
	store.RUnlock()

	store.Lock()
	store.Cloud.Config = config
	store.Cloud.Save()
	store.Unlock()

	store.runCallbacks("cloud")
}

func (store *Store) UpdateCloudState(state models.CloudState) {
	store.RLock()
	if store.Cloud.State.Equal(state) {
		store.RUnlock()
		return
	}
	store.RUnlock()

	store.Lock()
	store.Cloud.State = state
	store.Unlock()

	store.runCallbacks("cloud")
}
