package models

import "github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models/persons"

type ReadyInfo struct {
	Method string          `json:"method"`
	User   *persons.Person `json:"user"`
}
