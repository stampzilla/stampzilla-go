package models

import (
	"encoding/json"
	"fmt"
	"sync"
)

type DeviceState map[string]interface{}

func (ds DeviceState) Clone() DeviceState {
	newState := make(DeviceState)
	for k, v := range ds {
		newState[k] = v
	}
	return newState
}

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

// Copy copies a device
func (d *Device) Copy() *Device {
	d.Lock()
	newTraits := make([]string, len(d.Traits))
	newState := d.State.Clone()

	copy(newTraits, d.Traits)

	newD := &Device{
		Type:   d.Type,
		Node:   d.Node,
		ID:     d.ID,
		Name:   d.Name,
		Online: d.Online,
		State:  newState,
		Traits: d.Traits,
	}
	d.Unlock()
	return newD
}

type DeviceMap map[string]*Device

type Devices struct {
	devices DeviceMap
	sync.RWMutex
}

func NewDevices() *Devices {
	return &Devices{
		devices: make(DeviceMap),
	}
}

// Add adds a device to the list
func (d *Devices) Add(dev *Device) {
	d.Lock()
	d.devices[dev.Node+"."+dev.ID] = dev
	d.Unlock()
}

// Update the state of a device
func (d *Devices) SetState(node, id string, state DeviceState) error {
	d.Lock()
	defer d.Unlock()
	if dev, ok := d.devices[node+"."+id]; ok {
		dev.State = state
		return nil
	}

	return fmt.Errorf("Node not found (%s.%s)", node, id)
}

// Get returns a device
func (d *Devices) Get(node, id string) *Device {
	d.RLock()
	defer d.RUnlock()
	return d.devices[node+"."+id]
}

// All get all devices
func (d *Devices) All() DeviceMap {
	d.RLock()
	defer d.RUnlock()
	return d.devices
}

// Copy copies a list of devices
func (d *Devices) Copy() *Devices {

	newD := &Devices{
		devices: make(map[string]*Device),
	}
	d.RLock()
	for _, v := range d.devices {
		newD.Add(v.Copy())
	}
	d.RUnlock()

	return newD
}

func (d *Devices) MarshalJSON() ([]byte, error) {
	d.RLock()
	defer d.RUnlock()
	return json.Marshal(d.devices)
}

func (d *Devices) UnmarshalJSON(b []byte) error {
	var devices DeviceMap
	if err := json.Unmarshal(b, &devices); err != nil {
		return err
	}

	for _, dev := range devices {
		d.Add(dev)
	}
	return nil
}

// Flatten can be used for metrics export and logic rules
func (d *Devices) Flatten() map[string]interface{} {

	devmap := make(map[string]interface{})
	for k, v := range d.All() {
		v.Lock()
		for stateKey, s := range v.State {
			key := fmt.Sprintf("%s.%s", k, stateKey)
			devmap[key] = s
		}
		v.Unlock()
	}
	return devmap
}
