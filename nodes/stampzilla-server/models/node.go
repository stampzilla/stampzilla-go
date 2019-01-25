package models

import (
	"encoding/json"
	"sync"
)

type Node struct {
	UUID       string `json:"uuid,omitempty"`
	Connected_ bool   `json:"connected,omitempty"`
	Version    string `json:"version,omitempty"`
	Type       string `json:"type,omitempty"`
	Name       string `json:"name,omitempty"`
	//Devices   Devices         `json:"devices,omitempty"`
	Config json.RawMessage `json:"config,omitempty"`
	sync.Mutex
}

func (n *Node) SetConnected(c bool) {
	n.Lock()
	n.Connected_ = c
	n.Unlock()
}
func (n *Node) Connected() bool {
	n.Lock()
	defer n.Unlock()
	return n.Connected_
}
