package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-deconz/models"
)

func createUser() {
	client := &http.Client{Timeout: 5 * time.Second}
	u := fmt.Sprintf("http://%s:%s/api", config.IP, config.Port)
	req, err := http.NewRequest("POST", u, bytes.NewBufferString(`{"devicetype":"stampzilla"}`))
	req.SetBasicAuth("delight", config.Password)
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}

	decoder := json.NewDecoder(resp.Body)
	defer resp.Body.Close()
	type response []struct {
		Success struct {
			Username string `json:"username"`
		} `json:"success"`
	}

	data := response{}
	err = decoder.Decode(&data)
	if err != nil {
		log.Println(err)
		return
	}

	// bodyText, err := ioutil.ReadAll(resp.Body)
	// fmt.Println(string(bodyText))
	if len(data) != 1 {
		log.Println("Error wrong response")
		return
	}

	localConfig.APIKey = data[0].Success.Username
	err = localConfig.Save()
	if err != nil {
		logrus.Error(err)
	}
}

type API struct {
	Key    string
	Config *Config
	Client *http.Client
}

func NewAPI(key string, config *Config) *API {
	return &API{
		Key:    key,
		Client: &http.Client{Timeout: 5 * time.Second},
		Config: config,
	}
}

// ALL sensors
// GET /api/<apikey>/sensors
// All lights
// GET /api/<apikey>/lights

func (a *API) do(method, path string, body io.Reader, v interface{}) error {
	u := fmt.Sprintf("http://%s:%s/api/%s/%s", a.Config.IP, a.Config.Port, a.Key, path)
	req, err := http.NewRequest(method, u, body)
	if err != nil {
		return err
	}

	resp, err := a.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode > 399 {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("api status: %s error: %s", resp.Status, string(b))
	}
	logrus.Debug("status is ", resp.Status)

	if v == nil {
		return nil
	}
	decoder := json.NewDecoder(resp.Body)
	return decoder.Decode(v)
}

func (a *API) Get(path string, v interface{}) error {
	return a.do("GET", path, nil, v)
}

func (a *API) Post(path string, body io.Reader, v interface{}) error {
	return a.do("POST", path, body, v)
}

func (a *API) Put(path string, body io.Reader, v interface{}) error {
	return a.do("PUT", path, body, v)
}

func (a *API) PutData(path string, v interface{}) error {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	err := enc.Encode(v)
	if err != nil {
		return err
	}
	// TODO read response awsell and verify errors
	return a.do("PUT", path, &buf, nil)
}

func (a *API) Lights() (models.Lights, error) {
	lights := models.NewLights()
	err := a.Get("lights", &lights)
	return lights, err
}

func (a *API) Sensors() (models.Sensors, error) {
	data := models.NewSensors()
	err := a.Get("sensors", &data)
	return data, err
}
