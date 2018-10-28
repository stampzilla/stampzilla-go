package main

type Connection struct {
	Type       string `json:"type"`
	RemoteAddr string `json:"remoteAddr"`
	NodeUuid   string `json:"nodeUuid,omitEmpty"`
}
