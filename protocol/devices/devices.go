package devices

import (
	"encoding/json"
	"errors"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"

	log "github.com/cihub/seelog"
)

//type DeviceInterface interface {
//[> TODO: add methods <]
//SetName(string)
//SyncState(interface{})
//}

var regex = regexp.MustCompile(`^([^\s\[][^\s\[]*)?(\[.*?([0-9]+).*?\])?$`)

//Device is the abstraction between devices on the nodes and the device on the server
type Device struct {
	Type     string                 `json:"type"`
	Node     string                 `json:"node,omitempty"`
	Id       string                 `json:"id,omitempty"`
	Name     string                 `json:"name,omitempty"`
	Online   bool                   `json:"online"`
	StateMap map[string]string      `json:"stateMap,omitempty"`
	State    map[string]interface{} `json:"state,omitempty"`
	Tags     []string               `json:"tags"`

	sync.RWMutex
}

// NewDevice returns a new Device
func NewDevice() *Device {
	device := &Device{
		StateMap: make(map[string]string),
	}

	return device
}

//func (d *Device) SetName(name string) {
//d.Name = name
//}

// SyncState syncs the state between the node data and the device
func (d *Device) SyncState(state interface{}) {
	d.Lock()

	var err error
	d.State = make(map[string]interface{})
	for k, v := range d.StateMap {
		if value, err := path(state, v); err == nil {
			if reflect.ValueOf(value).IsNil() { // Dont accept nil values
				delete(d.State, k) // Remove it from the map
				continue
			}
			d.State[k] = value
			continue
		}
		log.Error(err)
	}
	d.StateMap = nil

	d.Unlock()
}

// SetOnline sets the online state of the device
func (d *Device) SetOnline(online bool) {
	d.Lock()
	d.Online = online
	d.Unlock()
}

// Map is a list of all devices. The key should be "<nodeuuid>.<deviceuuid>"
type Map struct {
	devices map[string]*Device
	sync.RWMutex
}

// NewMap returns initialized Map
func NewMap() *Map {
	return &Map{
		devices: make(map[string]*Device),
	}
}

// Add adds a device the the device map.
func (m *Map) Add(dev *Device) {
	m.Lock()
	defer m.Unlock()

	if dev.Node != "" {
		m.devices[dev.Node+"."+dev.Id] = dev
		return
	}
	m.devices[dev.Id] = dev
}

// Deletes a device from the map.
func (m *Map) Delete(id string) {
	m.Lock()
	defer m.Unlock()

	delete(m.devices, id)
}

// Exists return true if device id exists in the map
func (m *Map) Exists(id string) bool {
	m.RLock()
	defer m.RUnlock()

	if _, ok := m.devices[id]; ok {
		return true
	}
	return false
}

// ByID returns device based on id. If device has node set the id will be node.id
func (m *Map) ByID(uuid string) *Device {
	m.RLock()
	defer m.RUnlock()

	if node, ok := m.devices[uuid]; ok {
		return node
	}
	return nil
}

// All returns a list of all devices
func (m *Map) All() map[string]*Device {
	m.RLock()
	defer m.RUnlock()

	return m.devices
}

// Len returns length of map
func (m *Map) Len() int {
	m.RLock()
	defer m.RUnlock()

	return len(m.devices)
}

func (m *Map) MarshalJSON() ([]byte, error) {
	alias := make(map[string]Device)
	m.RLock()

	for k, v := range m.devices {
		if v != nil {
			alias[k] = *v
		}
	}
	defer m.RUnlock()
	return json.Marshal(alias)
}

func (m *Map) UnmarshalJSON(b []byte) error {
	var alias map[string]*Device
	err := json.Unmarshal(b, &alias)
	if err != nil {
		return err
	}

	m.Lock()
	m.devices = alias
	m.Unlock()

	return nil
}

func path(state interface{}, jp string) (interface{}, error) {
	if jp == "" {
		return nil, errors.New("invalid path")
	}
	for _, token := range strings.Split(jp, ".") {
		sl := regex.FindAllStringSubmatch(token, -1)
		if len(sl) == 0 {
			return nil, errors.New("invalid path1")
		}
		ss := sl[0]
		if ss[1] != "" {
			switch v1 := state.(type) {
			case map[string]interface{}:
				state = v1[ss[1]]
			}
		}
		if ss[3] != "" {
			ii, err := strconv.Atoi(ss[3])
			is := ss[3]
			if err != nil {
				return nil, errors.New("invalid path2")
			}
			switch v2 := state.(type) {
			case []interface{}:
				state = v2[ii]
			case map[string]interface{}:
				state = v2[is]
			}
		}
	}
	return state, nil
}
