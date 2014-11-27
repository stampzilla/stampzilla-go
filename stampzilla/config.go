package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
)

type Autostart struct {
	Name   string `json:"name"`
	Config string `json:"config"`
}
type Config struct {
	Autostart []*Autostart `json:"autostart"`
}

func (c *Config) generateDefault() error {
	nodes, err := ioutil.ReadDir("/home/stampzilla/go/src/github.com/stampzilla/stampzilla-go/nodes/")
	if err != nil {
		return err
	}

	config := &Config{}
	for _, node := range nodes {
		if !strings.Contains(node.Name(), "stampzilla-") {
			continue
		}
		name := strings.Replace(node.Name(), "stampzilla-", "", 1)
		autostart := &Autostart{Name: name, Config: "/var/spool/stampzilla/config/" + node.Name()}
		config.Autostart = append(config.Autostart, autostart)
	}

	*c = *config
	return nil
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
func (c *Config) readConfigFromFile(filepath string) error {
	configFile, err := os.Open(filepath)
	if err != nil {
		return err
	}

	config := &Config{}
	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&config); err != nil {
		return err
	}

	*c = *config
	return nil
}
