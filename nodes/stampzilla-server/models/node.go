package models

import (
	"encoding/json"
	"sync"

	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models/devices"
)

type Node struct {
	UUID       string `json:"uuid,omitempty"`
	Connected_ bool   `json:"connected,omitempty"`
	Version    string `json:"version,omitempty"`
	Type       string `json:"type,omitempty"`
	Name       string `json:"name,omitempty"`
	//Devices   Devices         `json:"devices,omitempty"`
	Config  json.RawMessage       `json:"config,omitempty"`
	Aliases map[devices.ID]string `json:"aliases,omitempty"`
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
