package protocol

import (
	"bytes"
	"encoding/json"
	"os"
	"sync"

	log "github.com/cihub/seelog"

	"github.com/stampzilla/stampzilla-go/protocol/devices"
)

type Devices struct {
	devices *devices.Map
	sync.RWMutex
}

func NewDevices() *Devices {
	n := &Devices{}
	n.devices = devices.NewMap()
	return n
}

func (n *Devices) ByUuid(uuid string) *devices.Device {
	n.RLock()
	defer n.RUnlock()
	return n.devices.ByID(uuid)
	return nil
}
func (n *Devices) All() map[string]*devices.Device {
	n.RLock()
	defer n.RUnlock()
	return n.devices.All()
}
func (n *Devices) ShallowCopy() map[string]*devices.Device {
	n.RLock()
	defer n.RUnlock()
	r := make(map[string]*devices.Device)
	for k, v := range n.devices.All() {
		copy := *v
		r[k] = &copy
	}
	return r
}
func (n *Devices) AllWithState(nodes *Nodes) map[string]*devices.Device {
	devices := n.ShallowCopy()
	for _, device := range devices {
		node := nodes.ByUuid(device.Node)
		if node == nil {
			device.State = nil
			continue //node is offline and we dont have the state
		}
		device.SyncState(node.State())
	}
	return devices
}
func (n *Devices) Add(device *devices.Device) error {
	n.Lock()
	defer n.Unlock()

	if dev := n.devices.ByID(device.Node + "." + device.Id); dev != nil {
		// Save name and tags
		device.Name = dev.Name
		device.Tags = dev.Tags
	}

	n.devices.Add(device)
	return nil
}
func (n *Devices) Delete(uuid string) {
	n.Lock()
	defer n.Unlock()
	n.devices.Delete(uuid)
}

//SetOfflineByNode marks device from a single node offline and returns a list of all marked devices
func (n *Devices) SetOfflineByNode(nodeUUID string) (list []*devices.Device) {
	n.Lock()
	defer n.Unlock()

	list = make([]*devices.Device, 0)
	for _, dev := range n.devices.All() {
		if dev.Node == nodeUUID {
			dev.SetOnline(false)
			list = append(list, dev)
		}
	}

	return
}

func (n *Devices) SaveToFile(path string) {
	configFile, err := os.Create(path)
	if err != nil {
		log.Error("creating config file", err.Error())
		return
	}
	var out bytes.Buffer
	b, err := json.Marshal(n.devices)
	if err != nil {
		log.Error("error marshal json", err)
	}
	json.Indent(&out, b, "", "\t")
	out.WriteTo(configFile)
}

func (n *Devices) RestoreFromFile(path string) {
	configFile, err := os.Open(path)
	if err != nil {
		log.Warn("opening config file", err.Error())
		return
	}

	type localDevice struct {
		Type string   `json:"type"`
		Node string   `json:"node"`
		ID   string   `json:"id"`
		Name string   `json:"name"`
		Tags []string `json:"tags"`
	}

	var devs map[string]*localDevice
	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&devs); err != nil {
		log.Error(err)
	}

	for _, v := range devs {
		n.Add(&devices.Device{
			Type:   v.Type,
			Node:   v.Node,
			Id:     v.ID,
			Online: false,
			Name:   v.Name,
			Tags:   v.Tags,
		})
	}
}
