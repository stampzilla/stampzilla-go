package main

type Config struct {
	Interval string
	Host     string
	Port     string
	Devices  []Device
}

type Device struct {
	Interval       string
	Enabled        bool
	Name           string `json:"keyName"`
	PrimaryAddress int
	// Frames which frames and datarecord to fetch.
	// Only 0 (first frame) is supported for now
	Frames map[string][]Record
}

type Record struct {
	Id   int
	Name string
}

func NewConfig() *Config {
	return &Config{}
}
