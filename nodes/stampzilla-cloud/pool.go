package main

import (
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
)

type Pool struct {
	Register   chan *Client
	Unregister chan *Client

	byID       map[string]*Client
	byInstance map[string]*Client
	sync.RWMutex
}

func NewPool() *Pool {
	return &Pool{
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		byID:       make(map[string]*Client),
		byInstance: make(map[string]*Client),
	}
}

func (pool *Pool) Start() {
	for {
		select {
		case client := <-pool.Register:
			pool.Lock()
			// Dont accept new connections that use the same client ID
			if _, ok := pool.byID[client.ID]; ok {
				pool.Unlock()
				go client.Disconnect()
				break
			}

			pool.byID[client.ID] = client
			pool.byInstance[client.Instance] = client
			pool.Unlock()
			logrus.Info("Size of Connection Pool: ", len(pool.byID))
			break
		case client := <-pool.Unregister:
			pool.Lock()
			delete(pool.byID, client.ID)
			delete(pool.byInstance, client.Instance)
			pool.Unlock()
			logrus.Info("Size of Connection Pool: ", len(pool.byID))
			break
		}
	}
}

func (pool *Pool) GetByID(id string) (*Client, error) {
	pool.RLock()
	defer pool.RUnlock()

	client, ok := pool.byID[id]
	if !ok {
		return nil, fmt.Errorf("instance not found")
	}

	return client, nil
}

func (pool *Pool) GetByInstance(instance string) (*Client, error) {
	pool.RLock()
	defer pool.RUnlock()

	client, ok := pool.byInstance[instance]
	if !ok {
		return nil, fmt.Errorf("instance not found")
	}

	return client, nil
}
