package store

import (
	"fmt"

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

	oldDest := store.Destinations.Get(dest.UUID)
	if oldDest != nil && oldDest.Equal(dest) {
		return
	}

	store.Destinations.Add(dest)
	store.Destinations.Save()
	store.runCallbacks("destinations")
}

func (store *Store) TriggerDestination(dest string, body string) error {
	destination := store.Destinations.Get(dest)
	if destination == nil {
		return fmt.Errorf("Destination defintion not found")
	}

	sender, ok := store.Senders.Get(destination.Sender)
	if !ok {
		return fmt.Errorf("Sender not found")
	}

	return sender.Trigger(destination, body)
}

func (store *Store) ReleaseDestination(dest string, body string) error {
	destination := store.Destinations.Get(dest)
	if destination == nil {
		return fmt.Errorf("Destination defintion not found")
	}

	sender, ok := store.Senders.Get(destination.Sender)
	if !ok {
		return fmt.Errorf("Sender not found")
	}

	return sender.Release(destination, body)
}

func (store *Store) GetSenderDestinations(id string) (error, map[string]string) {
	sender, ok := store.Senders.Get(id)
	if !ok {
		return fmt.Errorf("Sender not found"), nil
	}

	return sender.Destinations()
}
