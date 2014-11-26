package main

import (
	"bytes"
	"encoding/json"
	"os"
)

type Autostart struct {
	Name   string `json:"name"`
	config string `json:"config"`
}
type Config struct {
	Autostart []Autostart `json:"autostart"`
}

func (c *Config) SaveToFile(filepath string) error {
	configFile, err := os.Create(filepath)
	if err != nil {
		return err
	}

	var out bytes.Buffer
	b, err := json.MarshalIndent(c, "", "\t")
	if err != nil {
		return err
	}
	json.Indent(&out, b, "", "\t")
	out.WriteTo(configFile)
	return nil
}
func (c *Config) readConfigFromFile() error {
	configFile, err := os.Open("config.json")
	if err != nil {
		return err
	}

	config := &Config{}
	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&config); err != nil {
		return err
	}

	c = config
	return nil
}
