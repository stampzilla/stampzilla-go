package main

import "encoding/hex"

type State struct {
	Devices map[string]*Device
}

func NewState() *State {
	return &State{make(map[string]*Device)}
}

func (s *State) AddDevice(id [4]byte, name string, features []string, state string) {
	d := NewDevice(id, name, state, "", features)

	s.Devices[d.Id()] = d
}

func (s *State) GetState() interface{} {
	return s
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
}

func NewDevice(id [4]byte, name, state, dtype string, features []string) *Device {
	d := &Device{Name: name, State: state, Type: dtype}
	d.SetId(id)
	return d
}

func (d *Device) Id() string {
	return d.SenderId
}
func (d *Device) SetId(senderId [4]byte) {
	d.SenderId = hex.EncodeToString(senderId[0:4])

}
