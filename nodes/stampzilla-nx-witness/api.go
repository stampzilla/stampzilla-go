package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/sirupsen/logrus"
)

type API struct {
	config *Config
}

func NewAPI(config *Config) *API {
	return &API{
		config: config,
	}
}

func (api *API) fetchEventRules() (EventRulesResponse, error) {
	url := fmt.Sprintf("%s/ec2/getEventRules", api.config.Host)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("fetchEventRules: error create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(api.config.Username, api.config.Password)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("error fetching from api status: %d", resp.StatusCode)
	}

	response := EventRulesResponse{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (api *API) saveEventRule(rule Rule) error {
	url := fmt.Sprintf("%s/ec2/saveEventRule", api.config.Host)
	data, err := json.Marshal(&rule)
	if err != nil {
		return fmt.Errorf("saveEventRule: error marshal json: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("saveEventRule: error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(api.config.Username, api.config.Password)

	logrus.Debugf("req to url: %s", url)

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("saveEventRule: error making request: %w", err)
	}

	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("saveEventRule: error reading body: %w", err)
	}

	logrus.Debugf("statusCode: %s", resp.Status)
	logrus.Debugf("body: %s", b)

	if resp.StatusCode != 200 {
		return fmt.Errorf("error saving to api, status: %d, body: %s", resp.StatusCode, string(b))
	}

	return nil
}
