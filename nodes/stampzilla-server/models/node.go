package models

import (
	"encoding/json"
	"sync"

	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/models/devices"
	"github.com/stampzilla/stampzilla-go/v2/pkg/build"
)

type Node struct {
	UUID       string     `json:"uuid,omitempty"`
	Connected_ bool       `json:"connected,omitempty"`
	Version    build.Data `json:"version,omitempty"`
	Type       string     `json:"type,omitempty"`
	Name       string     `json:"name,omitempty"`
	// Devices   Devices         `json:"devices,omitempty"`
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

func (n *Node) SetAlias(id devices.ID, alias string) {
	n.Lock()
	if n.Aliases == nil {
		n.Aliases = make(map[devices.ID]string)
	}
	n.Aliases[id] = alias
	n.Unlock()
}

func (n *Node) Alias(id devices.ID) string {
	n.Lock()
	defer n.Unlock()
	if a, ok := n.Aliases[id]; ok {
		return a
	}
	return ""
}
