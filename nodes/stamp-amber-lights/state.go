package main

type State struct { /*{{{*/
	Devices []*Device
	Sensors map[string]string
} /*}}}*/


func (s *State) AddDevice(id, name string, features []string, state string) {
	d := NewDevice(id, name, state, "", features)

	s.Devices = append(s.Devices, d)
}

func (s *State) GetState() interface{} {
	return s;
}
