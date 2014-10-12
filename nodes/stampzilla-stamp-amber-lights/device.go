package main

type Device struct { /*{{{*/
	Id       string
	Name     string
	State    string
	Type     string
	Features []string
} /*}}}*/

func NewDevice(id, name, state, atype string, features []string) *Device {
	return &Device{id, name, state, atype, features}
}
