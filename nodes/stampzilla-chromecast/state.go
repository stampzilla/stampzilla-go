package main

import (
	"sync"

	"github.com/stampzilla/stampzilla-go/nodes/basenode"
	"github.com/stampzilla/stampzilla-go/protocol"
	"github.com/stampzilla/stampzilla-go/protocol/devices"
)

type State struct {
	Devices    map[string]*Chromecast
	connection basenode.Connection
	node       *protocol.Node
	sync.RWMutex
}

func (s *State) Publish() {
	if s.node != nil {
		s.connection.Send(s.node.Node())
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

	s.node.Devices_.Add(&devices.Device{
		Type:   "chromecast",
		Name:   c.Name_,
		Id:     c.Id,
		Online: true,
		Node:   s.node.Uuid(),
		StateMap: map[string]string{
			"Playing":    c.Id + ".Playing",
			"PrimaryApp": c.Id + ".PrimaryApp",
			"Title":      c.Id + ".Media.Title",
			"SubTitle":   c.Id + ".Media.SubTitle",
			"Thumb":      c.Id + ".Media.Thumb",
			"Url":        c.Id + ".Media.Url",
			"Duration":   c.Id + ".Media.Duration",
		},
	})

	s.Publish()
}

func (s *State) Remove(c *Chromecast) {
	_, ok := s.Devices[c.Uuid()]

	if ok {
		s.Lock()
		delete(s.Devices, c.Uuid())
		s.Unlock()
	}

	if dev := s.node.Devices_.ByID(c.Id); dev != nil {
		dev.Lock()
		dev.Online = false
		dev.Unlock()
	}

	s.Publish()
}
