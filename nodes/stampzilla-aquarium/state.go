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

	AirTemperature float64
	Humidity       float64

	WaterLevel float64

	Lights struct {
		White       float64
		Red         float64
		Green       float64
		Blue        float64
		Temperature float64
		Cooling     float64
	}

	PH float64
} /*}}}*/
