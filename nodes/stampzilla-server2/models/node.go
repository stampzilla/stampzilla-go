package models

type Node struct {
	Uuid      string            `json:"uuid"`
	Connected bool              `json:"connected"`
	Version   string            `json:"version"`
	Name      string            `json:"name"`
	State     interface{}       `json:"state"`
	WriteMap  map[string]bool   `json:"writeMap"`
	Config    map[string]string `json:"config"`
}
