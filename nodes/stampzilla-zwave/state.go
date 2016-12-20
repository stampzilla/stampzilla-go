package main

import (
	"strconv"
	"sync"

	"github.com/stampzilla/gozwave"
	"github.com/stampzilla/gozwave/nodes"
)

// Zwavenode is the node visible to the outside world through stampzilla-server
type Zwavenode struct {
	ID      int    `json:"id"`
	Brand   string `json:"brand"`
	Product string `json:"product"`
	Awake   bool   `json:"awake"`

	StateFloat map[string]float64 `json:"stateFloat"`
	StateBool  map[string]bool    `json:"stateBool"`

	node *nodes.Node
}

func newZwavenode(znode *nodes.Node) *Zwavenode {
	z := &Zwavenode{
		ID:         znode.Id,
		StateFloat: make(map[string]float64),
		StateBool:  make(map[string]bool),
		node:       znode,
	}

	return z
}

func (znode *Zwavenode) sync(node *nodes.Node) {
	if node.Device != nil {
		znode.Brand = node.Device.Brand
		znode.Product = node.Device.Product

		znode.StateBool = node.StateBool
		znode.StateFloat = node.StateFloat
	}
}

type State struct {
	Nodes map[string]*Zwavenode
	zwave *gozwave.Controller
	sync.Mutex
}

func NewState() *State {
	return &State{
		Nodes: make(map[string]*Zwavenode, 0),
	}
}

func (state *State) GetNode(address int) *Zwavenode {
	node, _ := state.Nodes[strconv.Itoa(address)]
	//for _, v := range state.Nodes {
	//if v.Id == address {
	//return v
	//}

	//}
	return node
}
