package store

import "github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models/notification"

func (store *Store) GetSenders() map[string]notification.Sender {
	store.RLock()
	defer store.RUnlock()
	return store.Senders.All()
}

func (store *Store) AddOrUpdateSender(sender notification.Sender) {
	if sender.UUID == "" {
		return
	}

	store.Senders.Add(sender)
	store.Senders.Save("senders.json")
	store.runCallbacks("senders")
}
