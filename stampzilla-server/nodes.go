package main

import (
	"encoding/json"
	"net"
	"net/http"
	"sync"
	"time"

	log "github.com/cihub/seelog"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/encoder"
	"github.com/stampzilla/stampzilla-go/protocol"
)

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

func (n *Nodes) GetByName(name string) *Node {
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
func (n *Nodes) GetByUuid(uuid string) *Node {
	n.RLock()
	defer n.RUnlock()
	if node, ok := n.nodes[uuid]; ok {
		return node
	}
	return nil
}
func (n *Nodes) GetAll() map[string]*Node {
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

var nodes *Nodes

func GetNodes(enc encoder.Encoder) (int, []byte) {
	return 200, encoder.Must(json.Marshal(nodes.GetAll()))
}

func GetNode(enc encoder.Encoder, params martini.Params) (int, []byte) {
	if n := nodes.GetByName(params["id"]); n != nil {
		return 200, encoder.Must(json.Marshal(&n))
	}
	if n := nodes.GetByUuid(params["id"]); n != nil {
		return 200, encoder.Must(json.Marshal(&n))
	}

	return 404, []byte("{}")
}

type Command struct {
	Cmd  string
	Args []string
}

func PostNodeState(enc encoder.Encoder, params martini.Params, r *http.Request) (int, []byte) {
	// Create a blocking channel
	nodesConnection[params["id"]].wait = make(chan bool)

	soc, ok := nodesConnection[params["id"]]
	if ok {
		//c := Command{}

		c := decodeJson(r)
		//err := r.DecodeJsonPayload(&c)

		data, err := json.Marshal(&c)

		_, err = soc.conn.Write(data)
		if err != nil {
			log.Error("Failed write: ", err)
		} else {
			log.Info("Success transport command to: ", params["id"])
		}
	} else {
		log.Error("Failed to transport command to: ", params["id"])
	}

	// Wait for answer or timeout..
	select {
	case <-time.After(5 * time.Second):
	case <-nodesConnection[params["id"]].wait:
	}

	n := nodes.GetByName(params["id"])
	if n == nil {
		return 404, []byte("{}")
	}

	//w.WriteJson(&n.State)
	return 200, encoder.Must(json.Marshal(&n.State))
}

func decodeJson(r *http.Request) interface{} {

	decoder := json.NewDecoder(r.Body)
	var t interface{}
	err := decoder.Decode(&t)
	if err != nil {
		log.Error(err)
	}
	return t
}

func CommandToNode(r *http.Request) {
	//  TODO: implement command here (jonaz) <Fri 03 Oct 2014 05:55:52 PM CEST>
}
