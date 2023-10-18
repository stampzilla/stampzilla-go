package main

import "time"

type Config struct {
	CarChargeDuration string `json:"carChargeDuration"`
	carChargeDuration time.Duration
	Token             string `json:"token"`
	HomeID            string `json:"homeId"`
}

func NewConfig() *Config {
	return &Config{}
}
