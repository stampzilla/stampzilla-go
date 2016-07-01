package main

import (
	"sync"

	"github.com/stampzilla/gozwave/nodes"
	"github.com/stampzilla/gozwave/serialapi"
)

type State struct {
	Nodes nodes.List
	zwave *serialapi.Connection
	sync.Mutex
}

func NewState() *State {
	return &State{}
}
