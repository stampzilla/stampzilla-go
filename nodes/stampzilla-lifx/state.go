package main

import (
	"encoding/hex"
	"sync"
)

type State struct {
	Lamps map[string]*Lamp
	sync.Mutex
}

type Lamp struct {
	Id   string
	Name string
	sync.Mutex
}

func NewState() *State {
	return &State{Lamps: make(map[string]*Lamp)}
}

func (s *State) Lamp(id [4]byte) *Lamp {
	senderId := hex.EncodeToString(id[0:4])
	s.Lock()
	defer s.Unlock()
	if _, ok := s.Lamps[senderId]; ok {
		return s.Lamps[senderId]
	}
	return nil
}
func (s *State) AddDevice(id, name string) *Lamp {
	d := NewLamp(id, name)
	s.Lock()
	defer s.Unlock()
	s.Lamps[d.Id] = d
	return d
}
func (s *State) RemoveDevice(id [4]byte) {
	s.Lock()
	defer s.Unlock()
	senderId := hex.EncodeToString(id[0:4])
	delete(s.Lamps, senderId)
}

func (s *State) GetState() interface{} {
	s.Lock()
	defer s.Unlock()
	return s
}

func NewLamp(id, name string) *Lamp {
	d := &Lamp{Name: name}
	d.SetId(id)
	return d
}

func (d *Lamp) SetId(id string) {
	d.Lock()
	defer d.Unlock()
	d.Id = id

}
