package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var config = &Config{}

type Config struct {
	IP       string
	Port     string
	Password string
}

var localConfig = &LocalConfig{}

type LocalConfig struct {
	APIKey string
}

func (lc LocalConfig) Save() error {
	path := "localconfig.json"
	configFile, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("error saving local config: %s", err.Error())
	}
	encoder := json.NewEncoder(configFile)
	encoder.SetIndent("", "\t")
	err = encoder.Encode(lc)
	return errors.Wrap(err, "error saving local config")
}

func (lc *LocalConfig) Load() error {
	path := "localconfig.json"
	logrus.Debug("loading local config from ", path)
	configFile, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			logrus.Warn(err)
			return nil
		}
		return fmt.Errorf("error loading local config: %s", err.Error())
	}

	jsonParser := json.NewDecoder(configFile)
	err = jsonParser.Decode(&lc)
	return errors.Wrap(err, "error loading local config")

	//TODO loop over rules and generate UUIDs if needed. If it was needed save the rules again
}

/*
Config to put into gui:
{
	"ip":"192.168.13.1",
	"port":"9042",
	"password":"password"
}

*/
