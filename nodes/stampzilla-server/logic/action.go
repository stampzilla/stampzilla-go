package logic

import serverprotocol "github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/protocol"

type Action interface {
	Run()
	Uuid() string
	Name() string
	SetNodes(*serverprotocol.Nodes)
}

type action struct {
	Commands []*command `json:"commands"`
	nodes    *serverprotocol.Nodes
	Uuid_    string `json:"uuid"`
	Name_    string `json:"name"`
}

func (a *action) Uuid() string {
	return a.Uuid_
}
func (a *action) Name() string {
	return a.Name_
}
func (a *action) Run() {
	for _, v := range a.Commands {
		v.Run()
	}
}
func (a *action) SetNodes(n *serverprotocol.Nodes) {
	a.nodes = n
}
