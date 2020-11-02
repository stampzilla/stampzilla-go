package main

import (
	"fmt"
	"strings"
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

			// Dont accept new connections that use the same instance
			instance := strings.ToLower(strings.TrimSpace(client.Instance))
			if _, ok := pool.byInstance[instance]; ok {
				pool.Unlock()
				go client.Disconnect()
				break
			}

			pool.byID[client.ID] = client
			pool.byInstance[instance] = client
			pool.Unlock()
			logrus.Info("Size of Connection Pool: ", len(pool.byID))
			break
		case client := <-pool.Unregister:
			pool.Lock()
			instance := strings.ToLower(strings.TrimSpace(client.Instance))
			delete(pool.byID, client.ID)
			delete(pool.byInstance, instance)
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

	client, ok := pool.byInstance[strings.ToLower(strings.TrimSpace(instance))]
	if !ok {
		return nil, fmt.Errorf("instance not found")
	}

	return client, nil
}
