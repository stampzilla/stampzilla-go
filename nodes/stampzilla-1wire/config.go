package main

type Config struct {
	Interval string
}

func NewConfig() *Config {
	return &Config{}
}
