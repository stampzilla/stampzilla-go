package main

import "strconv"

type State struct { /*{{{*/
	Devices map[string]*Device
	Sensors map[string]*Sensor
} /*}}}*/

func (s *State) AddDevice(id, name string, features []string, state DeviceState) {
	d := NewDevice(id, name, state, "", features)
	s.Devices[id] = d
}

func (s *State) AddSensor(id int) *Sensor {
	d := NewSensor(id)
	s.Sensors[strconv.Itoa(id)] = d
	return d
}
func (s *State) GetSensor(id int) *Sensor {
	if sens, ok := s.Sensors[strconv.Itoa(id)]; ok {
		return sens
	}
	return nil
}

func (s *State) GetDevice(id string) *Device {
	if d, ok := s.Devices[id]; ok {
		return d
	}
	return nil
}
