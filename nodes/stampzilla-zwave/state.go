package main

import (
	"sync"

	"github.com/stampzilla/gozwave"
	"github.com/stampzilla/gozwave/nodes"
)

type State struct {
	Nodes []nodes.Node
	zwave *gozwave.Controller
	sync.Mutex
}

func NewState() *State {
	return &State{
		Nodes: make([]nodes.Node, 0),
	}
}
