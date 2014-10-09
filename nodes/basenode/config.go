// Package main provides ...
package basenode

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"code.google.com/p/go-uuid/uuid"
	log "github.com/cihub/seelog"
)

var config = &Config{}
var host string
var port string

func init() {
	flag.StringVar(&host, "host", "localhost", "Server host/ip")
	flag.StringVar(&port, "port", "8282", "Server port")
}

func NewConfig() *Config {
	var config = &Config{}

	config.Host = host
	config.Port = host

	return config
}

func SetConfig(c *Config) {

	configFromFile := readConfigFromFile()
	c.Merge(configFromFile)

	if c.Uuid == "" {
		c.Uuid = uuid.New()
	}

	// Load logger
	logger, err := log.LoggerFromConfigAsFile("../logconfig.xml")
	if err != nil {
		panic(err)
	}
	log.ReplaceLogger(logger)

	config = c
	fmt.Println("config:", config)
	saveConfigToFile()
}

func saveConfigToFile() {
	configFile, err := os.Create("config.json")
	if err != nil {
		log.Error("creating config file", err.Error())
	}

	log.Info("Save config: ", config)
	var out bytes.Buffer
	b, err := json.MarshalIndent(config, "", "\t")
	if err != nil {
		log.Error("error marshal json", err)
	}
	json.Indent(&out, b, "", "\t")
	out.WriteTo(configFile)
}

func readConfigFromFile() *Config {
	configFile, err := os.Open("config.json")
	if err != nil {
		log.Error("opening config file", err.Error())
	}

	config := &Config{}
	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&config); err != nil {
		log.Error("parsing config file", err.Error())
	}

	return config
}

type Config struct {
	Host string
	Port string
	Uuid string
}

func (c *Config) GetUuid() string {
	return c.Uuid
}

func (c *Config) Merge(c2 *Config) {

	if c.Host != c2.Host && c.Host == "localhost" {
		c.Host = c2.Host
	}
	if c.Port != c2.Port && c.Port == "8282" {
		c.Port = c2.Port
	}

	c.Uuid = c2.Uuid

}
