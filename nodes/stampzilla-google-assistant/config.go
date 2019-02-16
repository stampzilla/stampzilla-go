package main

type Config struct {
	Port         string `json:"port"`
	ClientID     string `json:"clientID"`
	ClientSecret string `json:"clientSecret"`
	ProjectID    string `json:"projectID"`
	APIKey       string `json:"APIKey"`
}

func NewConfig() *Config {
	return &Config{}
}
