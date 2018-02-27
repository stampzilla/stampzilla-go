package main

import "sync"

type nodeSpecificConfig struct {
	Port         string     `json:"port"`
	ListenPort   string     `json:"listenPort"`
	ClientID     string     `json:"clientID"`
	ClientSecret string     `json:"clientSecret"`
	ProjectID    string     `json:"projectID"`
	APIKey       string     `json:"APIKey"`
	DevicesField deviceList `json:"devices"`
	sync.Mutex
}

func (ns *nodeSpecificConfig) Devices() deviceList {
	ns.Lock()
	defer ns.Unlock()
	return ns.DevicesField
}
func (ns *nodeSpecificConfig) Device(id string) *Device {
	ns.Lock()
	defer ns.Unlock()
	for _, v := range ns.DevicesField {
		if v.ID == id {
			return v
		}

	}
	return nil
}
func (ns *nodeSpecificConfig) AddDevice(dev *Device) {
	ns.Lock()
	defer ns.Unlock()
	ns.DevicesField = append(ns.DevicesField, dev)
}
