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
	sync.RWMutex
}

func (s *State) Publish() {
	if s.node != nil {
		(*s.connection).Send(s.node.Node())
	}
}

func (s *State) GetByUUID(uuid string) *Chromecast {
	s.RLock()
	defer s.RUnlock()
	if val, ok := s.Devices[uuid]; ok {
		return val
	}
	return nil
}

func (s *State) Add(c *Chromecast) {
	c.publish = s.Publish

	if s.Devices == nil {
		s.Devices = make(map[string]*Chromecast, 0)
	}

	s.Lock()
	s.Devices[c.Uuid()] = c
	s.Unlock()

	s.Publish()
}

func (s *State) Remove(c *Chromecast) {
	_, ok := s.Devices[c.Uuid()]

	if ok {
		s.Lock()
		delete(s.Devices, c.Uuid())
		s.Unlock()
	}

	s.Publish()
}
