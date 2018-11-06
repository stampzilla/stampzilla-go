package store

import (
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/models"
)

type Nodes map[string]*models.Node
type Connections map[string]*models.Connection

type Store struct {
	Nodes       Nodes
	Connections Connections
	onUpdate    []func(*Store) error
	sync.RWMutex
}

func New() *Store {
	return &Store{
		Nodes:       make(Nodes),
		Connections: make(Connections),
		onUpdate:    make([]func(*Store) error, 0),
	}
}

func (store *Store) AddOrUpdateNode(node *models.Node) {
	store.Lock()
	store.Nodes[node.Uuid] = node
	store.Unlock()

	for _, callback := range store.onUpdate {
		if err := callback(store); err != nil {
			logrus.Error("store: ", err)
		}
	}
}

func (store *Store) Connection(id string) *models.Connection {
	store.RLock()
	defer store.RUnlock()
	if conn, ok := store.Connections["foo"]; ok {
		return conn
	}
	return nil
}

func (store *Store) AddOrUpdateConnection(id string, c *models.Connection) {
	store.Lock()
	store.Connections[id] = c
	store.Unlock()

	for _, callback := range store.onUpdate {
		if err := callback(store); err != nil {
			logrus.Error("store: ", err)
		}
	}
}

func (store *Store) RemoveConnection(id string) {
	store.Lock()
	delete(store.Connections, id)
	store.Unlock()

	for _, callback := range store.onUpdate {
		if err := callback(store); err != nil {
			logrus.Error("store: ", err)
		}
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

func (store *Store) OnUpdate(callback func(*Store) error) {
	store.Lock()
	store.onUpdate = append(store.onUpdate, callback)
	store.Unlock()
}
