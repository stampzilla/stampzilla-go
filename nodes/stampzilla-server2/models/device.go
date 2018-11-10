package models

import "sync"

type DeviceState map[string]interface{}
type Device struct {
	Type   string      `json:"type"`
	Node   string      `json:"node,omitempty"`
	ID     string      `json:"id,omitempty"`
	Name   string      `json:"name,omitempty"`
	Online bool        `json:"online"`
	State  DeviceState `json:"state,omitempty"`
	Traits []string    `json:"traits"`
	sync.RWMutex
}

type Devices map[string]*Device

//TODO Devices should have a clone method. Usefull when we want to evauluate rules etc and send metrics to influxdb
