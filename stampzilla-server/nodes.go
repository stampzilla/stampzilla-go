package main

import (
	"net"
	"sync"

	"github.com/stampzilla/stampzilla-go/protocol"
)

//global variable to hold our nodes. Initialized in main
var nodes *Nodes

type Node struct {
	protocol.Node
	conn net.Conn
	wait chan bool
}

//  TODO: write tests for Nodes struct (jonaz) <Fri 10 Oct 2014 04:31:22 PM CEST>
type Nodes struct {
	nodes map[string]*Node
	sync.RWMutex
}

func NewNodes() *Nodes {
	n := &Nodes{}
	n.nodes = make(map[string]*Node)
	return n
}

func (n *Nodes) ByName(name string) *Node {
	n.RLock()
	defer n.RUnlock()
	for _, node := range n.nodes {
		//  TODO: change Id to Name (jonaz) <Fri 10 Oct 2014 04:29:23 PM CEST>
		if node.Id == name {
			return node
		}
	}
	return nil
}
func (n *Nodes) Search(nameoruuid string) *Node {
	if n := n.ByName(nameoruuid); n != nil {
		return n
	}
	if n := nodes.ByUuid(nameoruuid); n != nil {
		return n
	}
	return nil
}
func (n *Nodes) ByUuid(uuid string) *Node {
	n.RLock()
	defer n.RUnlock()
	if node, ok := n.nodes[uuid]; ok {
		return node
	}
	return nil
}
func (n *Nodes) All() map[string]*Node {
	n.RLock()
	defer n.RUnlock()
	return n.nodes
}
func (n *Nodes) Add(node *Node) {
	n.Lock()
	defer n.Unlock()
	//var newNode = node
	n.nodes[node.Uuid] = node
}
func (n *Nodes) Delete(uuid string) {
	n.Lock()
	defer n.Unlock()
	delete(n.nodes, uuid)
}
