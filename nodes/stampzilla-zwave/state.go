package main

import (
	"sync"

	"github.com/stampzilla/gozwave"
	"github.com/stampzilla/gozwave/nodes"
)

type Zwavenode struct {
	Id      int    `json:"id"`
	Brand   string `json:"brand"`
	Product string `json:"product"`
	Awake   bool   `json:"awake"`

	StateFloat map[string]float64
	StateBool  map[string]bool

	node *nodes.Node
}

func newZwavenode(znode *nodes.Node) *Zwavenode {
	z := &Zwavenode{
		Id:         znode.Id,
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
	}
}

type State struct {
	Nodes []*Zwavenode
	zwave *gozwave.Controller
	sync.Mutex
}

func NewState() *State {
	return &State{
		Nodes: make([]*Zwavenode, 0),
	}
}

func (state *State) GetNode(address int) *Zwavenode {
	for _, v := range state.Nodes {
		if v.Id == address {
			return v
		}

	}
	return nil
}
