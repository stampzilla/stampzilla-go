package main

type Config struct {
	Interval string
	Host     string
}

func NewConfig() *Config {
	return &Config{}
}
