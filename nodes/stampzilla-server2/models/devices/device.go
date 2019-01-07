package devices

import (
	"encoding/json"
	"fmt"
	"sync"
)

type State map[string]interface{}

func (ds State) Clone() State {
	newState := make(State)
	for k, v := range ds {
		newState[k] = v
	}
	return newState
}

func (ds State) Diff(right State) State {
	diff := make(State)
	for k, v := range ds {
		if v != right[k] {
			diff[k] = right[k]
		}
	}
	return diff
}

type Device struct {
	Type   string   `json:"type"`
	Node   string   `json:"node,omitempty"`
	ID     string   `json:"id,omitempty"`
	Name   string   `json:"name,omitempty"`
	Alias  string   `json:"alias,omitempty"`
	Online bool     `json:"online"`
	State  State    `json:"state,omitempty"`
	Traits []string `json:"traits"`
	sync.RWMutex
}

func (d *Device) SyncState(state State) {
	for k, v := range state {
		//d.Lock() TODO check if locking is needed
		d.State[k] = v
		//d.Unlock()
	}
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

type List struct {
	devices DeviceMap
	sync.RWMutex
}

func NewList() *List {
	return &List{
		devices: make(DeviceMap),
	}
}

// Add adds a device to the list
func (d *List) Add(dev *Device) {
	d.Lock()
	d.devices[dev.Node+"."+dev.ID] = dev
	d.Unlock()
}

// Update the state of a device
func (d *List) SetState(node, id string, state State) error {
	d.Lock()
	defer d.Unlock()
	if dev, ok := d.devices[node+"."+id]; ok {
		dev.State = state
		return nil
	}

	return fmt.Errorf("Node not found (%s.%s)", node, id)
}

// Get returns a device
func (d *List) Get(node, id string) *Device {
	d.RLock()
	defer d.RUnlock()
	return d.devices[node+"."+id]
}
func (d *List) GetUnique(id string) *Device {
	d.RLock()
	defer d.RUnlock()
	return d.devices[id]
}

// All get all devices
func (d *List) All() DeviceMap {
	d.RLock()
	defer d.RUnlock()
	return d.devices
}

// Copy copies a list of devices
func (d *List) Copy() *List {

	newD := &List{
		devices: make(map[string]*Device),
	}
	d.RLock()
	for _, v := range d.devices {
		newD.Add(v.Copy())
	}
	d.RUnlock()

	return newD
}

func (d *List) MarshalJSON() ([]byte, error) {
	d.RLock()
	defer d.RUnlock()
	return json.Marshal(d.devices)
}

func (d *List) UnmarshalJSON(b []byte) error {
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
func (d *List) Flatten() map[string]interface{} {

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
