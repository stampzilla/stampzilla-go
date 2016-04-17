package main

import (
	"log"
	"net"
	"net/url"
	"sync"
)

type State struct {
	Devices map[string]*Device
	sync.Mutex
}

func NewState() *State {
	return &State{Devices: make(map[string]*Device)}
}

func (s *State) Device(id string) *Device {
	s.Lock()
	defer s.Unlock()
	if _, ok := s.Devices[id]; ok {
		return s.Devices[id]
	}
	//TODO also check the name!
	return nil
}
func (s *State) AddDevice(d *Device) {
	s.Lock()
	defer s.Unlock()
	s.Devices[d.Id] = d
}
func (s *State) RemoveDevice(id string) {
	s.Lock()
	defer s.Unlock()
	delete(s.Devices, id)
}

func NewDevice(id, name string) *Device {
	d := &Device{Name: name}
	d.SetId(id)
	return d
}

type Device struct {
	Id      string
	Name    string
	Power   bool
	Playing bool
	Volume  int
	Ip      net.IP
	Title   string
	sync.Mutex
}

func (d *Device) SetId(id string) {
	d.Lock()
	defer d.Unlock()
	mac, err := net.ParseMAC(id)
	if err != nil {
		log.Println(err)
		return
	}

	d.Id = mac.String()
}
func (d *Device) IdUrlEncoded() string {
	return url.QueryEscape(d.Id)
}
