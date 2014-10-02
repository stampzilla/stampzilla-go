package protocol

type Node struct { /*{{{*/
	Id      string
	Actions []*Action
	Layout  []*Layout
	State   State
} /*}}}*/

type State interface {
	GetState() interface{}
}

func NewNode(name string, state State) *Node {
	return &Node{
		Id:      name,
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

func (n *Node) SetState(state State) {
	n.State = state
}
