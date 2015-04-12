package main

type State struct { /*{{{*/
	WaterTemperature float32
	Cooling          float32
	Filling          bool
	WaterLevelOk     bool

	Lights struct {
		White       float32
		Blue        float32
		Temperature float32
	}
} /*}}}*/

func (s *State) GetState() interface{} {
	return s
}
