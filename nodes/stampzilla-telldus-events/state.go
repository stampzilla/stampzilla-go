package main

type State struct { /*{{{*/
	Devices []*Device
	Sensors map[string]*Sensor
} /*}}}*/

func (s *State) AddDevice(id, name string, features []string, state DeviceState) {
	d := NewDevice(id, name, state, "", features)

	s.Devices = append(s.Devices, d)
}

func (s *State) GetState() interface{} {
	return s
}