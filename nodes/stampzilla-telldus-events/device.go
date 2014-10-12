package main

type Device struct { /*{{{*/
	Id       string
	Name     string
	State    DeviceState
	Type     string
	Features []string
} /*}}}*/

type DeviceState struct {
	On  bool
	Dim int
}

func NewDevice(id, name string, state DeviceState, atype string, features []string) *Device {
	return &Device{id, name, state, atype, features}
}
