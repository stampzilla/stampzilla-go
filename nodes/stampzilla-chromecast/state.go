package main

import (
	"sync"
)

type State struct {
	Chromecasts map[string]*Chromecast
	sync.RWMutex
}

func (s *State) GetByUUID(uuid string) *Chromecast {
	s.RLock()
	defer s.RUnlock()
	if val, ok := s.Chromecasts[uuid]; ok {
		return val
	}
	return nil
}

func (s *State) Add(c *Chromecast) {
	s.Lock()
	s.Chromecasts[c.Uuid()] = c
	s.Unlock()

	/*
		s.node.Chromecasts.Add(&devices.Device{
			Type:   "chromecast",
			Name:   c.Name(),
			Id:     c.Id,
			Online: true,
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
	*/
}

// Remove removes a chromecast from the state
func (s *State) Remove(c *Chromecast) {
	_, ok := s.Chromecasts[c.Uuid()]
	if ok {
		s.Lock()
		delete(s.Chromecasts, c.Uuid())
		s.Unlock()
	}
}
