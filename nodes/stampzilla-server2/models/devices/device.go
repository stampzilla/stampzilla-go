package devices

import (
	"encoding/json"
	"fmt"
	"strings"
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

// Bool runs fn only if key is found in map and it is of type bool
func (ds State) Bool(key string, fn func(bool)) {
	if v, ok := ds[key]; ok {
		if v, ok := v.(bool); ok {
			fn(v)
		}
	}
}

// Int runs fn only if key is found in map and it is of type int
func (ds State) Int(key string, fn func(int64)) {
	if v, ok := ds[key]; ok {
		if v, ok := v.(int); ok {
			fn(int64(v))
		}
		if v, ok := v.(int64); ok {
			fn(v)
		}
	}
}

// Float runs fn only if key is found in map and it is of type int
func (ds State) Float(key string, fn func(float64)) {
	if v, ok := ds[key]; ok {
		if v, ok := v.(float64); ok {
			fn(v)
		}
	}
}

//TODO diff function should handle new state on right side. Should compare length and do something smart
func (ds State) Diff(right State) State {
	diff := make(State)
	for k, v := range ds {
		if v != right[k] {
			diff[k] = right[k]
		}
	}

	for k, v := range right {
		if _, ok := ds[k]; !ok {
			diff[k] = v
		}
	}

	return diff
}

// Merge two states
func (ds State) Merge(right State) State {
	diff := make(State)
	for k, v := range ds {
		diff[k] = v
	}
	for k, v := range right {
		diff[k] = v
	}
	return diff
}

type ID struct {
	Node string
	ID   string
}

func (id ID) String() string {
	return id.Node + "." + id.ID
}

func (id ID) Bytes() []byte {
	return []byte(id.Node + "." + id.ID)
}
func (id ID) MarshalText() (text []byte, err error) {
	text = id.Bytes()
	return
}

func (id *ID) UnmarshalText(text []byte) error {
	tmp := strings.SplitN(string(text), ".", 2)
	if len(tmp) != 2 {
		return fmt.Errorf("wrong ID format. Expected nodeuuid.deviceid")
	}
	id.Node = tmp[0]
	id.ID = tmp[1]
	return nil

}

type Device struct {
	Type string `json:"type"`
	//Node   string   `json:"node,omitempty"`
	ID     ID       `json:"id,omitempty"`
	Name   string   `json:"name,omitempty"`
	Alias  string   `json:"alias,omitempty"`
	Online bool     `json:"online"`
	State  State    `json:"state,omitempty"`
	Traits []string `json:"traits"`
	sync.RWMutex
}

// Copy copies a device
func (d *Device) Copy() *Device {
	d.Lock()
	newTraits := make([]string, len(d.Traits))
	newState := d.State.Clone()

	copy(newTraits, d.Traits)

	newD := &Device{
		Type: d.Type,
		//Node: d.Node,
		ID: ID{
			ID:   d.ID.ID,
			Node: d.ID.Node,
		},
		Name:   d.Name,
		Online: d.Online,
		State:  newState,
		Traits: d.Traits,
	}
	d.Unlock()
	return newD
}

type DeviceMap map[ID]*Device

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
	d.devices[dev.ID] = dev
	d.Unlock()
}

// Update the state of a device
func (d *List) SetState(id ID, state State) error {
	d.Lock()
	defer d.Unlock()
	if dev, ok := d.devices[id]; ok {
		dev.State = state
		return nil
	}

	return fmt.Errorf("Node not found (%s)", id.String())
}

// Get returns a device
func (d *List) Get(id ID) *Device {
	d.RLock()
	defer d.RUnlock()
	return d.devices[id]
}

//func (d *List) GetUnique(id string) *Device {
//d.RLock()
//defer d.RUnlock()
//return d.devices[id]
//}

// All get all devices
func (d *List) All() DeviceMap {
	d.RLock()
	defer d.RUnlock()
	return d.devices
}

// Copy copies a list of devices
func (d *List) Copy() *List {

	newD := &List{
		devices: make(DeviceMap),
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
