package main

type Config struct {
	Host string
}

func NewConfig() *Config {
	return &Config{}
}
