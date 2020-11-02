package models

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
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

func (a CloudConfig) Equal(b CloudConfig) bool {
	if a.Enable != b.Enable {
		return false
	}
	if a.Server != b.Server {
		return false
	}
	if a.Instance != b.Instance {
		return false
	}

	return true
}

type CloudState struct {
	Secure    bool   `json:"secure"`
	Connected bool   `json:"connected"`
	Error     string `json:"error"`
}

func (a CloudState) Equal(b CloudState) bool {
	if a.Secure != b.Secure {
		return false
	}
	if a.Connected != b.Connected {
		return false
	}
	if a.Error != b.Error {
		return false
	}

	return true
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

type AuthorizeRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type ForwardedRequest struct {
	Dump       []byte `json:"dump"`
	URL        string `json:"url"`
	Service    string `json:"service"`
	RemoteAddr string `json:"remote_addr"`
}

func ParseForwardedRequest(raw json.RawMessage) (*ForwardedRequest, error) {
	fwd := &ForwardedRequest{}
	err := json.Unmarshal(raw, fwd)
	if err != nil {
		return nil, err
	}

	return fwd, nil
}

func (fr *ForwardedRequest) ParseRequest() (*http.Request, error) {
	b := bytes.NewBuffer(fr.Dump)
	rd := bufio.NewReader(b)
	return http.ReadRequest(rd)
}
