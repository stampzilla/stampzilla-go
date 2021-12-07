package main

import "time"

type Config struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Interval string `json:"interval"`
	interval time.Duration
}

func NewConfig() *Config {
	return &Config{}
}
