package protocol

import (
	"net"
	"sync"

	log "github.com/cihub/seelog"
	"github.com/stampzilla/stampzilla-go/protocol"
)

//global variable to hold our nodes. Initialized in main
var nodes *Nodes

type Node interface {
	SetState(interface{})
	State() interface{}
	Uuid() string
	Name() string
	SetUuid(string)
	SetName(string)
	Write(b []byte)
	SetConn(conn net.Conn)
}

type node struct {
	protocol.Node
	conn net.Conn
	wait chan bool
	sync.RWMutex
	//encoder *json.Encoder
}

func NewNode() Node {
	return &node{}
}

func (n *node) Conn() net.Conn {
	n.Lock()
	defer n.Unlock()
	return n.conn
}

func (n *node) SetConn(conn net.Conn) {
	n.Lock()
	n.conn = conn
	n.Host = conn.RemoteAddr().String()
	n.Unlock()
}
func (n *node) Write(b []byte) {
	b = append(b, []byte("\n")...)
	_, err := n.conn.Write(b)
	if err != nil {
		log.Error(err)
		return
	}
}

//  TODO: write tests for Nodes struct (jonaz) <Fri 10 Oct 2014 04:31:22 PM CEST>
type Nodes struct {
	nodes map[string]Node
	sync.RWMutex
}

func NewNodes() *Nodes {
	n := &Nodes{}
	n.nodes = make(map[string]Node)
	return n
}

func (n *Nodes) ByName(name string) Node {
	n.RLock()
	defer n.RUnlock()
	for _, node := range n.nodes {
		if node.Name() == name {
			return node
		}
	}
	return nil
}
func (n *Nodes) Search(nameoruuid string) Node {
	if n := n.ByUuid(nameoruuid); n != nil {
		return n
	}
	if n := n.ByName(nameoruuid); n != nil {
		return n
	}
	return nil
}
func (n *Nodes) ByUuid(uuid string) Node {
	n.RLock()
	defer n.RUnlock()
	if node, ok := n.nodes[uuid]; ok {
		return node
	}
	return nil
}
func (n *Nodes) All() map[string]Node {
	n.RLock()
	defer n.RUnlock()
	r := make(map[string]Node)
	for k, v := range n.nodes {
		r[k] = v
	}
	return r
	//return n.nodes
}
func (n *Nodes) Add(node Node) {
	n.Lock()
	defer n.Unlock()
	n.nodes[node.Uuid()] = node
}
func (n *Nodes) Delete(uuid string) {
	n.Lock()
	defer n.Unlock()
	delete(n.nodes, uuid)
}
