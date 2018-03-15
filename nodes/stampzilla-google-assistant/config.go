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

func newNodeSpecificConfig() *nodeSpecificConfig {
	return &nodeSpecificConfig{
		DevicesField: make(deviceList),
	}
}

func (ns *nodeSpecificConfig) Devices() deviceList {
	ns.Lock()
	defer ns.Unlock()
	return ns.DevicesField
}
func (ns *nodeSpecificConfig) Device(id string) *Device {
	ns.Lock()
	defer ns.Unlock()
	if v, ok := ns.DevicesField[id]; ok {
		return v
	}
	return nil
}
func (ns *nodeSpecificConfig) AddDevice(dev *Device) {
	ns.Lock()
	defer ns.Unlock()
	ns.DevicesField[dev.ID] = dev
}
