package main

import (
	"encoding/hex"
	"sync"
)

type State struct {
	Devices map[string]*Device
	sync.Mutex
}

func NewState() *State {
	return &State{Devices: make(map[string]*Device)}
}

func (s *State) Device(id [4]byte) *Device {
	senderId := hex.EncodeToString(id[0:4])
	s.Lock()
	defer s.Unlock()
	if _, ok := s.Devices[senderId]; ok {
		return s.Devices[senderId]
	}
	return nil
}
func (s *State) AddDevice(id, name string, state string) *Device {
	d := NewDevice(id, name, state, "")
	s.Lock()
	defer s.Unlock()
	s.Devices[d.Id] = d
	return d
}
func (s *State) RemoveDevice(id [4]byte) {
	s.Lock()
	defer s.Unlock()
	senderId := hex.EncodeToString(id[0:4])
	delete(s.Devices, senderId)
}

func (s *State) GetState() interface{} {
	s.Lock()
	defer s.Unlock()
	return s
}

func NewDevice(id, name, state, dtype string) *Device {
	d := &Device{Name: name, State: state, Type: dtype}
	d.SetId(id)
	return d
}

type Device struct {
	Id    string
	Name  string
	State string
	Type  string
	sync.Mutex
}

func (d *Device) SetId(id string) {
	d.Lock()
	defer d.Unlock()
	d.Id = id

}
