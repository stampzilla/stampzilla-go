package protocol

import (
	"encoding/json"
	"sync"

	"github.com/stampzilla/stampzilla-go/protocol/devices"
)

type Node struct {
	Name_     string `json:"Name"`
	Uuid_     string `json:"Uuid"`
	Host      string
	Version   string
	BuildDate string
	Elements_ []*Element  `json:"Elements"`
	State_    interface{} `json:"State"`
	Devices_  devices.Map `json:"Devices"`
	Config_   *ConfigMap  `json:"config"`
	sync.RWMutex
}

func NewNode(name string) *Node {
	n := &Node{
		Name_:     name,
		Version:   "-",
		Elements_: []*Element{},
		Devices_:  make(devices.Map),
	}

	n.Config_ = NewConfigMap(n)

	return n
}

func (n *Node) AddElement(el *Element) {
	n.Lock()
	n.Elements_ = append(n.Elements_, el)
	n.Unlock()
}

func (n *Node) SetState(state interface{}) {
	n.Lock()
	n.State_ = state
	n.Unlock()
}
func (n *Node) Devices() devices.Map {
	n.RLock()
	defer n.RUnlock()
	return n.Devices_
}
func (n *Node) SetDevices(devices devices.Map) {
	n.Lock()
	n.Devices_ = devices
	n.Unlock()
}

func (n *Node) Config() *ConfigMap {
	n.RLock()
	defer n.RUnlock()
	return n.Config_
}
func (n *Node) SetConfig(c *ConfigMap) {
	n.Lock()
	n.Config_ = c
	n.Unlock()
}

func (n *Node) State() interface{} {
	n.RLock()
	defer n.RUnlock()
	return n.State_
}
func (n *Node) SetElements(elements []*Element) {
	n.Lock()
	n.Elements_ = elements
	n.Unlock()
}
func (n *Node) Elements() []*Element {
	n.RLock()
	defer n.RUnlock()
	return n.Elements_
}
func (n *Node) Node() *Node {
	n.RLock()
	defer n.RUnlock()
	return n
}
func (n *Node) Uuid() string {
	n.RLock()
	defer n.RUnlock()
	return n.Uuid_
}
func (n *Node) Name() string {
	n.RLock()
	defer n.RUnlock()
	return n.Name_
}
func (n *Node) SetUuid(uuid string) {
	n.Lock()
	defer n.Unlock()
	n.Uuid_ = uuid
}
func (n *Node) SetName(name string) {
	n.Lock()
	defer n.Unlock()
	n.Name_ = name
}

func (n *Node) JsonEncode() (string, error) {
	n.RLock()
	ret, err := json.Marshal(n)
	n.RUnlock()
	if err != nil {
		return "", err
	}
	return string(ret), nil
}

type Identifiable interface {
	Uuid() string
}
