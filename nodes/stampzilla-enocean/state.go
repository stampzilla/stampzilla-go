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
func (s *State) DeviceByString(senderId string) *Device {
	s.Lock()
	defer s.Unlock()
	if _, ok := s.Devices[senderId]; ok {
		return s.Devices[senderId]
	}
	return nil
}

func (s *State) AddDevice(id [4]byte, name string, features []string, on bool) *Device {
	d := NewDevice(id, name, on, "", features)
	s.Lock()
	defer s.Unlock()
	s.Devices[d.IdString()] = d
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

func NewDevice(id [4]byte, name string, on bool, dtype string, features []string) *Device {
	d := &Device{Name: name, On: on, Type: dtype}
	d.SetId(id)
	return d
}

type Device struct {
	SenderId  string
	Name      string
	On        bool
	Type      string
	Features  []string
	SendEEPs  []string
	RecvEEPs  []string
	Power     int64
	PowerUnit string
	sync.Mutex
}

func (d *Device) AddEepForSending(eep string) {
	d.Lock()
	defer d.Unlock()
	d.SendEEPs = append(d.SendEEPs, eep)
}
func (d *Device) AddEepForReceiving(eep string) {
	d.Lock()
	defer d.Unlock()
	d.RecvEEPs = append(d.RecvEEPs, eep)
}

func (d *Device) Id() [4]byte {
	d.Lock()
	defer d.Unlock()
	senderid, _ := hex.DecodeString(d.SenderId)
	var ret [4]byte
	copy(ret[:], senderid[0:4])
	return ret
}
func (d *Device) IdString() string {
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

func (d *Device) handler() Handler {
	return handlers.getHandler(d.SendEEPs[0])
}
func (d *Device) CmdOn() {
	d.handler().On(d)
}
func (d *Device) CmdOff() {
	d.handler().Off(d)
}
func (d *Device) CmdToggle() {
	d.handler().Toggle(d)
}
func (d *Device) CmdDim(val int) {
	d.handler().Dim(val, d)
}
