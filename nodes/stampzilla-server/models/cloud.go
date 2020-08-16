package models

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/sirupsen/logrus"
)

type Cloud struct {
	Config CloudConfig `json:"config"`
	State  CloudState  `json:"state"`
	sync.RWMutex
}

type CloudConfig struct {
	Enable   bool   `json:"enable"`
	Server   string `json:"server"`
	Instance string `json:"instance"`
}

type CloudState struct {
	Secure    bool   `json:"secure"`
	Connected bool   `json:"connected"`
	Error     string `json:"error"`
}

func (c *Cloud) Save() error {
	configFile, err := os.Create("cloud.json")
	if err != nil {
		return fmt.Errorf("cloud config: error saving: %s", err)
	}
	encoder := json.NewEncoder(configFile)
	encoder.SetIndent("", "\t")
	c.Lock()
	defer c.Unlock()
	err = encoder.Encode(c.Config)
	if err != nil {
		return fmt.Errorf("cloud config: error saving: %s", err)
	}
	return nil
}

func (c *Cloud) Load() error {
	logrus.Info("Loading cloud config from json file")

	configFile, err := os.Open("cloud.json")
	if err != nil {
		if os.IsNotExist(err) {
			logrus.Warn(err)
			return nil // We dont want to error our if the file does not exist when we start the server
		}
		return err
	}

	c.Lock()
	defer c.Unlock()
	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&c.Config); err != nil {
		return fmt.Errorf("persons: error loading: %s", err)
	}

	return nil
}
