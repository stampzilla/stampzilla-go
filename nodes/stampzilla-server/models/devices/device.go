package devices

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

type ID struct {
	Node string
	ID   string
}

// ErrWrongIDFormat is returned if the new ID is not in the correct format.
var ErrWrongIDFormat = fmt.Errorf("wrong ID format. Expected nodeuuid.deviceid")

func NewIDFromString(id string) (ID, error) {
	tmp := strings.SplitN(string(id), ".", 2)
	if len(tmp) != 2 {
		return ID{}, ErrWrongIDFormat
	}
	return ID{Node: tmp[0], ID: tmp[1]}, nil
}

func (id ID) String() string {
	return strings.Join([]string{id.Node, id.ID}, ".")
}

func (id ID) IsZero() bool {
	return id == ID{}
}

func (id ID) Bytes() []byte {
	return []byte(strings.Join([]string{id.Node, id.ID}, "."))
}

func (id ID) MarshalText() (text []byte, err error) {
	text = id.Bytes()
	return
}

func (id *ID) UnmarshalText(text []byte) error {
	tmp := strings.SplitN(string(text), ".", 2)
	if len(tmp) != 2 {
		return ErrWrongIDFormat
	}
	id.Node = tmp[0]
	id.ID = tmp[1]
	return nil
}

type Device struct {
	Type   string   `json:"type"`
	ID     ID       `json:"id"`
	Name   string   `json:"name,omitempty"`
	Alias  string   `json:"alias,omitempty"`
	Online bool     `json:"online"`
	State  State    `json:"state"`
	Traits []string `json:"traits"`
	sync.RWMutex
}

// NewDevice initializes a new device with needed maps.
func NewDevice() *Device {
	return &Device{
		State: make(State),
	}
}

// SetOnline set online state.
func (d *Device) SetOnline(v bool) {
	d.Lock()
	d.Online = v
	d.Unlock()
}

func (d *Device) MarshalJSON() ([]byte, error) {
	d.RLock()
	defer d.RUnlock()
	return json.Marshal(Device(*d))
}

// Equal checks if 2 devices are equal.
func (d *Device) Equal(dev *Device) bool {
	d.RLock()
	defer d.RUnlock()
	if !d.State.Equal(dev.State) { // this is most likely to not be equal so we check it first
		return false
	}
	if d.Type != dev.Type {
		return false
	}
	if d.Name != dev.Name {
		return false
	}
	if d.Alias != dev.Alias {
		return false
	}
	if d.Online != dev.Online {
		return false
	}
	if d.Online != dev.Online {
		return false
	}

	if !sliceIsEqual(d.Traits, dev.Traits) {
		return false
	}

	return true
}

func sliceIsEqual(a, b []string) bool {
	// If one is nil, the other must also be nil.
	if (a == nil) != (b == nil) {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

// Copy copies a device.
func (d *Device) Copy() *Device {
	d.Lock()
	newTraits := make([]string, len(d.Traits))
	newState := d.State.Clone()

	copy(newTraits, d.Traits)

	newD := &Device{
		Type: d.Type,
		// Node: d.Node,
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

// Len return how many devices there are in the list.
func (d *List) Len() int {
	d.Lock()
	defer d.Unlock()
	return len(d.devices)
}

// Add adds a device to the list.
func (d *List) Add(dev *Device) {
	d.Lock()
	d.devices[dev.ID] = dev
	d.Unlock()
}

// Update the state of a device.
func (d *List) SetState(id ID, state State) error {
	d.Lock()
	defer d.Unlock()
	if dev, ok := d.devices[id]; ok {
		dev.State = state
		return nil
	}

	return fmt.Errorf("Node not found (%s)", id.String())
}

// Get returns a device.
func (d *List) Get(id ID) *Device {
	d.RLock()
	defer d.RUnlock()
	return d.devices[id]
}

// func (d *List) GetUnique(id string) *Device {
// d.RLock()
// defer d.RUnlock()
// return d.devices[id]
//}

// All get all devices.
func (d *List) All() DeviceMap {
	d.RLock()
	defer d.RUnlock()
	return d.devices
}

func (d *List) Remove(id ID) {
	d.Lock()
	delete(d.devices, id)
	d.Unlock()
}

// Copy copies a list of devices.
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

// Flatten can be used for metrics export and logic rules.
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

// StateGroupedByNode get state from all devices grouped by node uuid.
func (d *List) StateGroupedByNode() map[string]map[ID]State {
	d.RLock()
	devicesByNode := make(map[string]map[ID]State)
	for id, state := range d.devices {
		if devicesByNode[id.Node] == nil {
			devicesByNode[id.Node] = make(map[ID]State)
		}
		devicesByNode[id.Node][id] = state.State
	}
	d.RUnlock()
	return devicesByNode
}
