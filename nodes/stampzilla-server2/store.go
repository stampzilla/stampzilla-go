package main

import (
	"sync"
)

type Nodes map[string]*Node
type Connections map[string]*Connection

type Store struct {
	Nodes       Nodes
	Connections Connections
	onUpdate    []func(*Store)
	sync.RWMutex
}

func NewStore() *Store {
	return &Store{
		Nodes:       make(Nodes),
		Connections: make(Connections),
		onUpdate:    make([]func(*Store), 0),
	}
}

func (store *Store) AddOrUpdateNode(node *Node) {
	store.Lock()
	store.Nodes[node.Uuid] = node
	store.Unlock()

	for _, callback := range store.onUpdate {
		callback(store)
	}
}

func (store *Store) AddOrUpdateConnection(id string, c *Connection) {
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

func (store *Store) GetNode(uuid string) *Node {
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
