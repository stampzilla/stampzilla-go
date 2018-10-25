package main

import (
	"sync"
)

type Nodes map[string]*Node

type Store struct {
	Nodes Nodes
	sync.RWMutex
}

type Node struct {
	Uuid      string            `json:"uuid"`
	Connected bool              `json:"connected"`
	Version   string            `json:"version"`
	Name      string            `json:"name"`
	State     interface{}       `json:"state"`
	WriteMap  map[string]bool   `json:"writeMap"`
	Config    map[string]string `json:"config"`
}

func NewStore() *Store {
	return &Store{
		Nodes: make(Nodes),
	}
}

func (store *Store) AddOrUpdateNode(node *Node) {
	store.Lock()
	store.Nodes[node.Uuid] = node
	store.Unlock()
}

func (store *Store) GetNodes() Nodes {
	store.RLock()
	defer store.RUnlock()
	return store.Nodes
}

func (store *Store) GetNode(uuid string) *Node {

	return nil
}
