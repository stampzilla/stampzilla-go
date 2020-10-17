package main

import (
	"encoding/json"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-nibe/nibe"
)

type Config struct {
	Port string `json:"port"`
}

func updatedConfig(n *nibe.Nibe) func(data json.RawMessage) error {
	return func(data json.RawMessage) error {
		logrus.Info("Received config from server:", string(data))

		newConf := &Config{}
		err := json.Unmarshal(data, newConf)
		if err != nil {
			return err
		}

		config = newConf
		logrus.Info("Config is now: ", config)

		n.Connect(config.Port)

		return nil
	}
}
