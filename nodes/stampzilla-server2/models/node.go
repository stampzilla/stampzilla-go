package models

import "encoding/json"

type Node struct {
	UUID      string          `json:"uuid,omitempty"`
	Connected bool            `json:"connected,omitempty"`
	Version   string          `json:"version,omitempty"`
	Type      string          `json:"type,omitempty"`
	Name      string          `json:"name,omitempty"`
	Devices   Devices         `json:"devices,omitempty"`
	Config    json.RawMessage `json:"config,omitempty"`
}
