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

// Copy copies a device
func (d *Device) Copy() *Device {
	newState := make(DeviceState)
	d.Lock()
	newTraits := make([]string, len(d.Traits))
	for k, v := range d.State {
		newState[k] = v
	}

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
	devices map[string]*Device
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

// Add adds a device to the list
func (d *Devices) Get(id string) *Device {
	d.RLock()
	defer d.RUnlock()
	return d.devices[id]
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
