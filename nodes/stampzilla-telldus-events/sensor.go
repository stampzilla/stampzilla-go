package main

type Sensor struct { /*{{{*/
	Id       string
	Name     string
	State    string
} /*}}}*/

func NewSensor(id, name, state string) *Sensor {
	return &Sensor{id, name, state}
}
