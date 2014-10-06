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
func (s *State) AddDevice(id [4]byte, name string, features []string, state string) *Device {
	d := NewDevice(id, name, state, "", features)
	s.Lock()
	defer s.Unlock()
	s.Devices[d.Id()] = d
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

func NewDevice(id [4]byte, name, state, dtype string, features []string) *Device {
	d := &Device{Name: name, State: state, Type: dtype}
	d.SetId(id)
	return d
}

type Device struct {
	SenderId  string
	Name      string
	State     string
	Type      string
	Features  []string
	EEPs      []string
	Power     int64
	PowerUnit string
	sync.Mutex
}

func (d *Device) AddEep(eep string) {
	d.Lock()
	defer d.Unlock()
	d.EEPs = append(d.EEPs, eep)
}

func (d *Device) Id() string {
	d.Lock()
	defer d.Unlock()
	return d.SenderId
}
func (d *Device) SetId(senderId [4]byte) {
	d.Lock()
	defer d.Unlock()
	d.SenderId = hex.EncodeToString(senderId[0:4])

}

func (d *Device) SetPower(pwr int64) {
	d.Lock()
	defer d.Unlock()
	d.Power = pwr
}

func (d *Device) GetPower() int64 {
	d.Lock()
	defer d.Unlock()
	return d.Power
}
