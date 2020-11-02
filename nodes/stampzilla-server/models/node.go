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
	Config       json.RawMessage                  `json:"config,omitempty"`
	Aliases      map[devices.ID]string            `json:"aliases,omitempty"`
	DeviceLabels map[devices.ID]map[string]string `json:"labels,omitempty"`
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

func (n *Node) SetLabels(id devices.ID, labels map[string]string) {
	n.Lock()
	if n.DeviceLabels == nil {
		n.DeviceLabels = make(map[devices.ID]map[string]string)
	}
	n.DeviceLabels[id] = labels
	n.Unlock()
}
func (n *Node) Labels(id devices.ID) map[string]string {
	n.Lock()
	defer n.Unlock()
	if a, ok := n.DeviceLabels[id]; ok {
		return a
	}
	return make(map[string]string)
}
