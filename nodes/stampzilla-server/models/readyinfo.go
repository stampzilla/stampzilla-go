package models

import "github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/models/persons"

type ReadyInfo struct {
	Method string          `json:"method"`
	User   *persons.Person `json:"user"`
}
