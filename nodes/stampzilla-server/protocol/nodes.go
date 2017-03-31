package protocol

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"

	log "github.com/cihub/seelog"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/notifications"
	"github.com/stampzilla/stampzilla-go/protocol"
	"github.com/stampzilla/stampzilla-go/protocol/devices"
)

//global variable to hold our nodes. Initialized in main
var nodes *Nodes

type Node interface {
	SetState(interface{})
	State() interface{}

	Devices() *devices.Map
	SetDevices(*devices.Map)

	Config() *protocol.ConfigMap
	SetConfig(*protocol.ConfigMap)

	SetElements([]*protocol.Element)
	Elements() []*protocol.Element

	Uuid() string
	Name() string
	SetUuid(string)
	SetName(string)
	Write(b []byte) (int, error)
	SetConn(conn net.Conn)
	GetNotification(json.RawMessage) *notifications.Notification
}

type node struct {
	//Notification *json.RawMessage
	Ping bool `json:",omitempty"`
	Pong bool `json:",omitempty"`

	protocol.Node
	conn net.Conn
	wait chan bool
	sync.RWMutex
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
func (n *node) Write(b []byte) (int, error) {
	b = append(b, []byte("\n")...)
	count, err := n.conn.Write(b)
	if err != nil {
		n.conn.Close()
		log.Error(err)
		return count, err
	}

	return count, nil
}

func (n *node) GetNotification(notification json.RawMessage) *notifications.Notification {
	if notification == nil {
		return nil
	}

	note := &notifications.Notification{}
	err := json.Unmarshal(notification, note)

	note.SourceUuid = n.Uuid()
	note.Source = n.Name()

	if err != nil {
		log.Warn("Failed to decode notification: ", err)
		return nil
	}

	return note
}

//  TODO: write tests for Nodes struct (jonaz) <Fri 10 Oct 2014 04:31:22 PM CEST>
type Nodes struct {
	nodes map[string]Node
	sync.RWMutex
}

type Searchable interface {
	Search(string) Node
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
func (n *Nodes) Add(node Node) error {
	n.Lock()
	defer n.Unlock()

	if node.Uuid() == "" {
		return fmt.Errorf("Failed to add node, does not contain an UUID")
	}

	n.nodes[node.Uuid()] = node

	return nil
}
func (n *Nodes) Delete(uuid string) {
	n.Lock()
	defer n.Unlock()
	delete(n.nodes, uuid)
}

func WriteUpdate(w io.Writer, msg *protocol.Update) error {
	bytes, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = w.Write(bytes)
	return err
}
