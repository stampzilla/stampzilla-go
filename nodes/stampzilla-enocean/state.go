package main

import (
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/Sirupsen/logrus"
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

func NewDevice(id [4]byte, name string, on bool, dtype string, features []string) *Device {
	d := &Device{Name: name, On_: on, Type: dtype}
	d.SetId(id)
	return d
}

type Device struct {
	SenderId string
	UniqueId int64
	Name     string
	On_      bool `json:"On"`
	Type     string
	Features []string
	SendEEPs []string
	RecvEEPs []string
	PowerW   int64
	PowerkWh int64
	Dim      int64
	Status   string
	sync.RWMutex
}

func (d *Device) HasSingleRecvEEP(eep string) bool {
	d.RLock()
	defer d.RUnlock()
	for _, v := range d.RecvEEPs {
		if v == eep {
			return true
		}

	}
	return false
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
	d.RLock()
	defer d.RUnlock()
	senderid, _ := hex.DecodeString(d.SenderId)
	var ret [4]byte
	copy(ret[:], senderid[0:4])
	return ret
}
func (d *Device) IdString() string {
	d.RLock()
	defer d.RUnlock()
	return d.SenderId
}
func (d *Device) SetOn(s bool) {
	d.Lock()
	d.On_ = s
	d.Unlock()
}
func (d *Device) On() bool {
	d.RLock()
	defer d.RUnlock()
	return d.On_
}
func (d *Device) SetId(senderId [4]byte) {
	d.Lock()
	defer d.Unlock()
	d.SenderId = hex.EncodeToString(senderId[0:4])

}

func (d *Device) SetPowerW(pwr int64) {
	d.Lock()
	defer d.Unlock()
	d.PowerW = pwr
}

func (d *Device) GetPowerW() int64 {
	d.RLock()
	defer d.RUnlock()
	return d.PowerW
}
func (d *Device) SetPowerkWh(pwr int64) {
	d.Lock()
	defer d.Unlock()
	d.PowerkWh = pwr
}

func (d *Device) GetPowerkWh() int64 {
	d.RLock()
	defer d.RUnlock()
	return d.PowerkWh
}

func (d *Device) handler() (Handler, error) {
	if len(d.SendEEPs) < 1 {
		return nil, fmt.Errorf("No SendEEPs defined on device %s", d.IdString())
	}
	return handlers.getHandler(d.SendEEPs[0]), nil
}
func (d *Device) CmdOn() {
	h, err := d.handler()
	if err != nil {
		logrus.Error(err)
		return
	}

	h.On(d)
}
func (d *Device) CmdOff() {
	h, err := d.handler()
	if err != nil {
		logrus.Error(err)
		return
	}
	h.Off(d)
}
func (d *Device) CmdToggle() {
	h, err := d.handler()
	if err != nil {
		logrus.Error(err)
		return
	}
	h.Toggle(d)
}
func (d *Device) CmdDim(val int) {
	h, err := d.handler()
	if err != nil {
		logrus.Error(err)
		return
	}
	h.Dim(val, d)
}
func (d *Device) CmdLearn() {
	h, err := d.handler()
	if err != nil {
		logrus.Error(err)
		return
	}
	h.Learn(d)
}
