package main

type Config struct {
	Weather struct {
		Current struct {
			Temperature struct {
				Device string `json:"device"`
				Field  string `json:"field"`
			} `json:"temperature"`
			Humidity struct {
				Device string `json:"device"`
				Field  string `json:"field"`
			} `json:"humidity"`
		} `json:"current"`
	} `json:"weather"`
}
