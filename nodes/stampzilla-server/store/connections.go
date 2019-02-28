package store

import "github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models"

func (store *Store) GetConnections() Connections {
	store.RLock()
	defer store.RUnlock()
	return store.Connections
}

func (store *Store) Connection(id string) *models.Connection {
	store.RLock()
	defer store.RUnlock()
	if conn, ok := store.Connections[id]; ok {
		return conn
	}
	return nil
}

func (store *Store) AddOrUpdateConnection(id string, c *models.Connection) {
	store.Lock()
	store.Connections[id] = c
	store.Unlock()

	store.runCallbacks("connections")
}

func (store *Store) ConnectionChanged() {
	store.runCallbacks("connections")
}

func (store *Store) RemoveConnection(id string) {
	store.Lock()
	delete(store.Connections, id)
	store.Unlock()
	store.runCallbacks("connections")
}
