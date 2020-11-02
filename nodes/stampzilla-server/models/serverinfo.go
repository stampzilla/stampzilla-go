package models

type ServerInfo struct {
	Name       string `json:"name"`
	UUID       string `json:"uuid"`
	TLSPort    string `json:"tlsPort"`
	Port       string `json:"port"`
	Init       bool   `json:"init"`
	AllowLogin bool   `json:"allowLogin"`

	Instance string `json:"instance,omitempty"`
	Phrase   string `json:"phrase,omitempty"`
}
