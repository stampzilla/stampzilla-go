package store

import (
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models/notification"
)

func (store *Store) GetDestinations() map[string]*notification.Destination {
	store.RLock()
	defer store.RUnlock()
	return store.Destinations.All()
}

func (store *Store) AddOrUpdateDestination(dest *notification.Destination) {
	if dest == nil {
		return
	}

	oldDest := store.Destinations.Get(dest.ID)
	if oldDest != nil && oldDest.Equal(dest) {
		return
	}

	store.Destinations.Add(dest)
	store.runCallbacks("destinations")
}
