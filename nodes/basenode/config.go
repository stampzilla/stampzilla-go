// Package main provides ...
package basenode

import (
	log "github.com/cihub/seelog"
)

var config = &Config{}

func SetConfig(c *Config) {

	// Load logger
	logger, err := log.LoggerFromConfigAsFile("../logconfig.xml")
	if err != nil {
		panic(err)
	}
	log.ReplaceLogger(logger)

	config = c
}

type Config struct {
	Host string
	Port string
}
