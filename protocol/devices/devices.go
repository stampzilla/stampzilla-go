package devices

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

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

	var err error
	d.State = make(map[string]interface{})
	for k, v := range d.StateMap {
		if value, err := path(state, v); err == nil {
			d.State[k] = value
			continue
		}
		log.Error(err)
	}
	d.StateMap = nil
}

// Map is a list of all devices. The key should be "<nodeuuid>.<deviceuuid>"
type Map map[string]*Device

// Add adds a device the the device map. Its not safe for concurrent use.
func (m Map) Add(dev *Device) {
	m[dev.Id] = dev
	//map[string]*Device(*m)[dev.Id] = dev
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
