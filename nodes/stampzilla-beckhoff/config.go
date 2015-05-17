package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"

	log "github.com/cihub/seelog"
)

var config = &Config{}
var ip string
var netid string
var port int
var tpy string

func init() {
	flag.StringVar(&ip, "ads-ip","","the address to the AMS router")
    flag.StringVar(&netid, "ads-netid","","AMS NetID of the target")
    flag.IntVar(&port, "ads-port",801,"AMS Port of the target")
    flag.StringVar(&tpy, "tpy","","Xml program description file (.tpy)")
}

func NewConfig() *Config {
	var config = &Config{}

	config.Ip = ip
	config.Netid = netid
	config.Port = port
	config.Tpy = tpy

	return config
}

func SetConfig(c *Config) {

	configFromFile := readConfigFromFile()
	c.Merge(configFromFile)

	config = c

	saveConfigToFile()
}

func saveConfigToFile() {
	configFile, err := os.Create("beckhoff.json")
	if err != nil {
		log.Error("creating config file", err.Error())
	}

	log.Info("Save beckhoff config: ", config)
	var out bytes.Buffer
	b, err := json.MarshalIndent(config, "", "\t")
	if err != nil {
		log.Error("error marshal json", err)
	}
	json.Indent(&out, b, "", "\t")
	out.WriteTo(configFile)
}

func readConfigFromFile() *Config {
	configFile, err := os.Open("beckhoff.json")
	if err != nil {
		log.Error("opening config file", err.Error())
	}

	config := &Config{}
	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&config); err != nil {
		log.Error("parsing beckhoff config file", err.Error())
	}

	return config
}

type Config struct {
	Ip string
	Netid string
	Port int
	Tpy string
}

func (c *Config) Merge(c2 *Config) {

	if c2.Ip != "" {
		c.Ip = c2.Ip
	}
	if c2.Netid != "" {
		c.Netid = c2.Netid
	}
	if c2.Port != 801 {
		c.Port = c2.Port
	}
	if c2.Tpy != "" {
		c.Tpy = c2.Tpy
	}

}
