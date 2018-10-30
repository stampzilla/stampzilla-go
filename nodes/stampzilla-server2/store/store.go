package store

import (
	"sync"

	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/models"
)

type Nodes map[string]*models.Node
type Connections map[string]*models.Connection

type Store struct {
	Nodes       Nodes
	Connections Connections
	onUpdate    []func(*Store)
	sync.RWMutex
}

func New() *Store {
	return &Store{
		Nodes:       make(Nodes),
		Connections: make(Connections),
		onUpdate:    make([]func(*Store), 0),
	}
}

func (store *Store) AddOrUpdateNode(node *models.Node) {
	store.Lock()
	store.Nodes[node.Uuid] = node
	store.Unlock()

	for _, callback := range store.onUpdate {
		callback(store)
	}
}

func (store *Store) AddOrUpdateConnection(id string, c *models.Connection) {
	store.Lock()
	store.Connections[id] = c
	store.Unlock()

	for _, callback := range store.onUpdate {
		callback(store)
	}
}

func (store *Store) RemoveConnection(id string) {
	store.Lock()
	delete(store.Connections, id)
	store.Unlock()

	for _, callback := range store.onUpdate {
		callback(store)
	}
}

func (store *Store) GetNodes() Nodes {
	store.RLock()
	defer store.RUnlock()
	return store.Nodes
}

func (store *Store) GetNode(uuid string) *models.Node {
	return nil
}

func (store *Store) GetConnections() Connections {
	store.RLock()
	defer store.RUnlock()
	return store.Connections
}

func (store *Store) OnUpdate(callback func(*Store)) {
	store.Lock()
	store.onUpdate = append(store.onUpdate, callback)
	store.Unlock()
}
