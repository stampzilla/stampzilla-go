package main

type Config struct {
	Interval string   `json:"interval"`
	Host     string   `json:"host"`
	Port     string   `json:"port"`
	Devices  []Device `json:"devices"`
}

type Device struct {
	Interval       string `json:"interval"`
	Enabled        bool   `json:"enabled"`
	Name           string `json:"name"`
	PrimaryAddress int    `json:"primaryAddress"`
	// Frames which frames and datarecord to fetch.
	// Only 0 (first frame) is supported for now
	Frames map[string][]Record `json:"frames"`
}

type Record struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func NewConfig() *Config {
	return &Config{}
}
