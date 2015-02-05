package main

type State struct { /*{{{*/
	Devices []*Device
	Sensors map[string]*Sensor
} /*}}}*/

func (s *State) AddDevice(id, name string, features []string, state DeviceState) {
	d := NewDevice(id, name, state, "", features)

	s.Devices = append(s.Devices, d)
}

func (s *State) GetDevice(id string) *Device {
	for _, v := range s.Devices {
		if v.Id == id {
			return v
		}

	}
	return nil
}

func (s *State) GetState() interface{} {
	return s
}
