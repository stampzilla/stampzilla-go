package main

type Sensor struct { /*{{{*/
	Id       int
	Name     string
	Temp     float64
	Humidity float64
} /*}}}*/

func NewSensor(id int, name string) *Sensor {
	return &Sensor{Id: id, Name: name}
}
