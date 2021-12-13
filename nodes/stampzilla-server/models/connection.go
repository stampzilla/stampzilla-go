package models

import "github.com/lesismal/melody"

type Connection struct {
	Type       string                 `json:"type"`
	RemoteAddr string                 `json:"remoteAddr"`
	NodeUuid   string                 `json:"nodeUuid,omitEmpty"`
	Attributes map[string]interface{} `json:"attributes"`

	Session *melody.Session `json:"-"`
}
