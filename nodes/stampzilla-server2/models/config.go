package models

import (
	"bytes"
	"encoding/json"
	"os"

	"github.com/google/uuid"
	"github.com/koding/multiconfig"
	"github.com/sirupsen/logrus"
)

//Config is the main server configuration
type Config struct {
	Port    string `json:"port" default:"8080"`
	TLSPort string `json:"tlsPort" default:"6443"`
	UUID    string `json:"uuid"`
	Name    string `json:"name"`
	Host    string `json:"host,omitempty"`
}

//Save writes the config as json to specified filename
func (c *Config) Save(filename string) {
	configFile, err := os.Create(filename)
	if err != nil {
		logrus.Error("creating config file", err.Error())
	}

	logrus.Info("Save config: ", c)
	var out bytes.Buffer
	b, err := json.MarshalIndent(c, "", "\t")
	if err != nil {
		logrus.Error("error marshal json", err)
	}
	json.Indent(&out, b, "", "\t")
	out.WriteTo(configFile)
}

//MustLoad loads the config using multiconfig from json or environment or command line args
func (c *Config) MustLoad() {
	m := c.createMultiConfig()
	m.MustLoad(c)

	if c.UUID == "" {
		c.UUID = uuid.New().String()
	}
}

func (c *Config) createMultiConfig() *multiconfig.DefaultLoader {
	loaders := []multiconfig.Loader{}

	// Read default values defined via tag fields "default"
	loaders = append(loaders, &multiconfig.TagLoader{})

	if _, err := os.Stat("config.json"); err == nil {
		loaders = append(loaders, &multiconfig.JSONLoader{Path: "config.json"})
	}

	e := &multiconfig.EnvironmentLoader{}
	e.Prefix = "STAMPZILLA"
	f := &multiconfig.FlagLoader{}
	f.EnvPrefix = "STAMPZILLA"

	loaders = append(loaders, e, f)
	loader := multiconfig.MultiLoader(loaders...)

	d := &multiconfig.DefaultLoader{}
	d.Loader = loader
	d.Validator = multiconfig.MultiValidator(&multiconfig.RequiredValidator{})
	return d

}
