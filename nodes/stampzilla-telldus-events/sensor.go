package main

type Sensor struct { /*{{{*/
	Id       int
	Temp     float64
	Humidity float64
} /*}}}*/

func NewSensor(id int) *Sensor {
	return &Sensor{Id: id}
}
