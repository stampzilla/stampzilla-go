package main

type State struct { /*{{{*/
	WaterTemperature float64
	Cooling          float64
	Heating          bool

	Skimmer          bool
	CirculationPumps bool

	Filling     bool
	FillingTime float64

	WaterLevelOk bool
	FilterOk     bool

	Lights struct {
		White       float32
		Red         float32
		Greed       float32
		Blue        float32
		Temperature float32
	}
} /*}}}*/
