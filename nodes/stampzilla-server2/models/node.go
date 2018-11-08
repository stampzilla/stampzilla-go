package models

import "encoding/json"

type Node struct {
	UUID      string            `json:"uuid"`
	Connected bool              `json:"connected"`
	Version   string            `json:"version"`
	Type      string            `json:"type"`
	Name      string            `json:"name"`
	Devices   map[string]Device `json:"devices"`
	Config    json.RawMessage   `json:"config"`
}
