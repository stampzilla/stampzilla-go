package protocol

import (
	"encoding/json"
	"sync"
)

type Node struct { /*{{{*/
	Name     string
	Uuid     string
	Host     string
	Actions  []*Action
	Layout   []*Layout
	Elements []*Element
	State_   interface{} `json:"State"`
	sync.RWMutex
} /*}}}*/

//type State interface {
//GetState() interface{}
//}

func NewNode(name string) *Node {
	return &Node{
		Name:    name,
		Uuid:    "",
		Host:    "",
		Actions: []*Action{},
		Layout:  []*Layout{}}
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
	n.Elements = append(n.Elements, el)
	n.Unlock()
}

func (n *Node) SetState(state interface{}) {
	n.Lock()
	n.State_ = state
	n.Unlock()
}
func (n *Node) State() interface{} {
	n.RLock()
	defer n.RUnlock()
	return n.State_
}
func (n *Node) Node() *Node {
	n.RLock()
	defer n.RUnlock()
	return n
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
