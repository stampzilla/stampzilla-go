package main

import (
	"sync"

	"github.com/stampzilla/stampzilla-go/nodes/basenode"
	"github.com/stampzilla/stampzilla-go/protocol"
)

type State struct {
	Devices    map[string]*Chromecast
	connection *basenode.Connection
	node       *protocol.Node
	sync.Mutex
}

func (s *State) Publish() {
	if s.node != nil {
		(*s.connection).Send(s.node.Node())
	}
}

func (s *State) Add(c *Chromecast) {
	c.publish = s.Publish

	if s.Devices == nil {
		s.Devices = make(map[string]*Chromecast, 0)
	}

	s.Lock()
	s.Devices[c.Name()] = c
	s.Unlock()

	s.Publish()
}

func (s *State) Remove(c *Chromecast) {
	_, ok := s.Devices[c.Name()]

	if ok {
		s.Lock()
		delete(s.Devices, c.Name())
		s.Unlock()
	}

	s.Publish()
}
