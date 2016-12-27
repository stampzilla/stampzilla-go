package main

import (
	"encoding/json"
	"time"

	log "github.com/cihub/seelog"
	"github.com/davecgh/go-spew/spew"
	resty "gopkg.in/resty.v0"
)

type LifxCloudClient struct {
	ApiAccessToken string
}

func NewLifxCloudClient(aat string) (*LifxCloudClient, error) {
	log.Info("create poller")
	lc := &LifxCloudClient{
		ApiAccessToken: aat,
	}

	return lc, nil
}

func (lc *LifxCloudClient) Start() {
	log.Info("Start poller")
	go lc.Poller()
}

func (lc *LifxCloudClient) Poller() {
	resp, err := resty.R().
		SetHeader("Accept", "application/json").
		SetAuthToken(lc.ApiAccessToken).
		Get("https://api.lifx.com/v1/lights/all")

	if err != nil {
		log.Errorf("Lifx cloud poller failed: %#v", err.Error())
		return
	}

	var data []*cloudGetAllResponse

	err = json.Unmarshal(resp.Body(), &data)
	if err != nil {
		log.Errorf("Lifx cloud poller failed decode: %#v", err.Error())
		return
	}

	spew.Dump(data)
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
