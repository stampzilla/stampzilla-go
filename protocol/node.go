package protocol

import "encoding/json"

type Node struct { /*{{{*/
	Name     string
	Uuid     string
	Actions  []*Action
	Layout   []*Layout
	Elements []*Element
	State    interface{}
} /*}}}*/

type State interface {
	GetState() interface{}
}

func NewNode(name string) *Node {
	return &Node{
		Name:    name,
		Uuid:    "",
		Actions: []*Action{},
		Layout:  []*Layout{}}
}

func (n *Node) AddAction(id, name string, args []string) {
	a := NewAction(id, name, args)

	n.Actions = append(n.Actions, a)
}

func (n *Node) AddLayout(id, atype, action, using string, filter []string, section string) {
	l := NewLayout(id, atype, action, using, filter, section)

	n.Layout = append(n.Layout, l)
}

func (n *Node) AddElement(el *Element) {
	n.Elements = append(n.Elements, el)
}

func (n *Node) SetState(state State) {
	n.State = state
}

func (n *Node) JsonEncode() (string, error) {
	ret, err := json.Marshal(n)
	if err != nil {
		return "", err
	}
	return string(ret), nil
}
