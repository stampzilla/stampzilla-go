package main

type Connection struct {
	Type       string                 `json:"type"`
	RemoteAddr string                 `json:"remoteAddr"`
	NodeUuid   string                 `json:"nodeUuid,omitEmpty"`
	Attributes map[string]interface{} `json:"attributes"`
}
