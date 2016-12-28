package main

import (
	"encoding/json"
	"fmt"
	"time"

	log "github.com/cihub/seelog"
	resty "gopkg.in/resty.v0"
)

type LifxCloudClient struct {
	ApiAccessToken   string
	pollerResultFunc func([]*cloudGetAllResponse)
}

func NewLifxCloudClient(aat string) (*LifxCloudClient, error) {
	log.Info("create poller")
	lc := &LifxCloudClient{
		ApiAccessToken:   aat,
		pollerResultFunc: func([]*cloudGetAllResponse) {},
	}

	return lc, nil
}

func (lc *LifxCloudClient) Start() {
	log.Info("Start poller")
	go lc.Poller()
}

func (lc *LifxCloudClient) Poller() {
	for {
		lamps, err := lc.Poll()
		if err != nil {
			log.Errorf("Poller failed: %s", err)
			<-time.After(60 * time.Second)
			continue
		}

		lc.pollerResultFunc(lamps)

		<-time.After(5 * time.Second)
	}
}

func (lc *LifxCloudClient) Poll() ([]*cloudGetAllResponse, error) {
	resp, err := resty.R().
		SetHeader("Accept", "application/json").
		SetAuthToken(lc.ApiAccessToken).
		Get("https://api.lifx.com/v1/lights/all")

	if err != nil {
		return nil, fmt.Errorf("Lifx cloud poller failed: %#v", err.Error())

	}

	var data []*cloudGetAllResponse

	err = json.Unmarshal(resp.Body(), &data)
	if err != nil {
		return nil, fmt.Errorf("Lifx cloud poller failed decode: %#v", err.Error())
	}

	return data, nil

	//spew.Dump(data)
}

type cloudGetAllResponse struct {
	ID         string        `json:"id"`
	Uuid       string        `json:"uuid"`
	Label      string        `json:"label"`
	Connected  bool          `json:"connected"`
	Power      string        `json:"power"`
	Color      cloudColor    `json:"color"`
	Infrared   string        `json:"infrared"`
	Brightness float64       `json:"brightness"`
	Group      cloudGroup    `json:"group"`
	Location   cloudLocation `json:"location"`
	LastSeen   time.Time     `json:"last_seen"`
	Product    cloudProduct  `json:"product"`
}

type cloudColor struct {
	Hue        float64 `json:"hue"`
	Saturation float64 `json:"saturation"`
	Kelvin     int     `json:"kelvin"`
}

type cloudGroup struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type cloudLocation struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type cloudProduct struct {
	Name        string          `json:"name"`
	Company     string          `json:"company"`
	Identifier  string          `json:"identifier"`
	Capabilites map[string]bool `json:"capabilites"`
}
