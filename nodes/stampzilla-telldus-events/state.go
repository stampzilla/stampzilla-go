package main

type State struct { /*{{{*/
	Devices map[string]*Device
	Sensors map[int]*Sensor
} /*}}}*/

func (s *State) AddDevice(id, name string, features []string, state DeviceState) {
	d := NewDevice(id, name, state, "", features)
	s.Devices[id] = d
}

func (s *State) AddSensor(id int, name string) *Sensor {
	d := NewSensor(id, name)
	s.Sensors[id] = d
	return d
}
func (s *State) GetSensor(id int) *Sensor {
	if sens, ok := s.Sensors[id]; ok {
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

func (s *State) GetState() interface{} {
	return s
}
