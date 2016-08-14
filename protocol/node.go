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
	Actions   []*Action
	Layout    []*Layout
	Elements_ []*Element  `json:"Elements"`
	State_    interface{} `json:"State"`
	Devices_  devices.Map `json:"Devices"`
	sync.RWMutex
}

func NewNode(name string) *Node {
	return &Node{
		Name_:     name,
		Actions:   []*Action{},
		Elements_: []*Element{},
		Layout:    []*Layout{},
		Devices_:  make(devices.Map),
	}
}

func (n *Node) AddAction(id, name string, args []string) {
	a := NewAction(id, name, args)

	n.Lock()
	n.Actions = append(n.Actions, a)
	n.Unlock()
}

func (n *Node) AddLayout(id, atype, action, using string, filter []string, section string) {
	l := NewLayout(id, atype, action, using, filter, section)

	n.Lock()
	n.Layout = append(n.Layout, l)
	n.Unlock()
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
