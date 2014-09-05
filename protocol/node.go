package protocol

type Node struct { /*{{{*/
	Id      string
	Actions []*Action
	Layout  []*Layout
	State   State
} /*}}}*/

func NewNode(name string) *Node {
	return &Node{
		name,
		[]*Action{},
		[]*Layout{},
		State{},
	}
}

func (n *Node) AddAction(id, name string, args []string) {
	a := NewAction(id, name, args)

	n.Actions = append(n.Actions, a)
}

func (n *Node) AddLayout(id, atype, action, using string, filter []string, section string) {
	l := NewLayout(id, atype, action, using, filter, section)

	n.Layout = append(n.Layout, l)
}

func (n *Node) AddDevice(id, name string, features []string, state string) {
	d := NewDevice(id, name, state, "", features)

	n.State.Devices = append(n.State.Devices, d)
}
