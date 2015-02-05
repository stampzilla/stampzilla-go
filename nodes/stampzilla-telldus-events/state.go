package main

type State struct { /*{{{*/
	Devices map[string]*Device
	Sensors map[string]*Sensor
} /*}}}*/

func (s *State) AddDevice(id, name string, features []string, state DeviceState) {
	d := NewDevice(id, name, state, "", features)

	s.Devices[id] = d
}

func (s *State) GetDevice(id string) *Device {
	if d, ok := s.Devices[id]; ok {
		return d
	}
	return nil
}

func (s *State) GetState() interface{} {
	return s
}
