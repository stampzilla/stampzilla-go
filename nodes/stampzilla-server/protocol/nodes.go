package protocol

import (
	"net"
	"sync"

	log "github.com/cihub/seelog"
	"github.com/stampzilla/stampzilla-go/protocol"
)

//global variable to hold our nodes. Initialized in main
var nodes *Nodes

type Node struct {
	protocol.Node
	conn net.Conn
	wait chan bool
	sync.RWMutex
	//encoder *json.Encoder
}

func (n *Node) Conn() net.Conn {
	n.Lock()
	defer n.Unlock()
	return n.conn
}

func (n *Node) SetConn(conn net.Conn) {
	n.Lock()
	n.conn = conn
	n.Host = conn.RemoteAddr().String()
	n.Unlock()
}
func (n *Node) Write(b []byte) {
	b = append(b, []byte("\n")...)
	_, err := n.conn.Write(b)
	if err != nil {
		log.Error(err)
		return
	}
}

//  TODO: write tests for Nodes struct (jonaz) <Fri 10 Oct 2014 04:31:22 PM CEST>
type Nodes struct {
	nodes []*Node
	sync.RWMutex
}

func NewNodes() *Nodes {
	n := &Nodes{}
	n.nodes = make([]*Node, 0)
	return n
}

func (n *Nodes) ByName(name string) *Node {
	n.RLock()
	defer n.RUnlock()
	for _, node := range n.nodes {
		if node.Name == name {
			return node
		}
	}
	return nil
}
func (n *Nodes) Search(nameoruuid string) *Node {
	if n := n.ByUuid(nameoruuid); n != nil {
		return n
	}
	if n := n.ByName(nameoruuid); n != nil {
		return n
	}
	return nil
}
func (n *Nodes) ByUuid(uuid string) *Node {
	n.RLock()
	defer n.RUnlock()

	for _, node := range n.nodes {
		if node.Uuid == uuid {
			return node
		}
	}

	return nil
}
func (n *Nodes) All() []*Node {
	n.RLock()
	defer n.RUnlock()
	return n.nodes
}
func (n *Nodes) Add(node *Node) {
	n.Lock()
	defer n.Unlock()
	n.nodes = append(n.nodes, node)
}
func (n *Nodes) Delete(uuid string) {
	n.Lock()
	defer n.Unlock()
	for i, node := range n.nodes {
		if node.Uuid == uuid {
			n.nodes = append(n.nodes[:i], n.nodes[i+1:]...)
			return
		}
	}
}
