package main

import "sync"

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
	return &Store{}
}

func (store *Store) AddOrUpdateNode(node *Node) {

	return
}

func (store *Store) GetNodes() Nodes {

	return store.Nodes
}

func (store *Store) GetNode(uuid string) *Node {

	return nil
}
