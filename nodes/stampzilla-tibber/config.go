package main

import "time"

const (
	ResolutionQuarterHourly = "QUARTER_HOURLY"
	ResolutionHourly        = "HOURLY"
)

type Config struct {
	CarChargeDuration string `json:"carChargeDuration"`
	carChargeDuration time.Duration
	Token             string `json:"token"`
	HomeID            string `json:"homeId"`
	// Resolution is the price resolution, QUARTER_HOURLY (default) or HOURLY.
	Resolution string `json:"resolution"`
	slot       time.Duration
}

func NewConfig() *Config {
	return &Config{}
}
