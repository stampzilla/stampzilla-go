package main

type Config struct {
	Interval  string
	Host      string
	Variables []Variables
}

type Variables struct {
	Name    string
	Address string
	Type    string
}

func NewConfig() *Config {
	return &Config{}
}
