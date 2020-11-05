package store

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/models/notification"
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
	err := store.Destinations.Save("destinations.json")
	if err != nil {
		logrus.Error(err)
		return
	}
	store.runCallbacks("destinations")
}

func (store *Store) TriggerDestination(dest string, body string) error {
	destination := store.Destinations.Get(dest)
	if destination == nil {
		return fmt.Errorf("destination definition not found")
	}

	sender, ok := store.Senders.Get(destination.Sender)
	if !ok {
		return fmt.Errorf("sender not found")
	}

	return sender.Trigger(destination, body)
}

func (store *Store) ReleaseDestination(dest string, body string) error {
	destination := store.Destinations.Get(dest)
	if destination == nil {
		return fmt.Errorf("destination definition not found")
	}

	sender, ok := store.Senders.Get(destination.Sender)
	if !ok {
		return fmt.Errorf("sender not found")
	}

	return sender.Release(destination, body)
}

func (store *Store) GetSenderDestinations(id string) (map[string]string, error) {
	sender, ok := store.Senders.Get(id)
	if !ok {
		return nil, fmt.Errorf("sender not found")
	}

	return sender.Destinations()
}
