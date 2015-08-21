package servernode

import (
	"net"
	"strconv"

	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/notifications"
	serverprotocol "github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/protocol"
	"github.com/stampzilla/stampzilla-go/protocol"
)

type Node struct {
	protocol.Node
	State_       map[string]interface{} `json:"State"`
	logicChannel chan string
}

func New(uuid string, logicChannel chan string) serverprotocol.Node {
	node := &Node{
		State_:       make(map[string]interface{}),
		logicChannel: logicChannel,
	}
	node.SetName("server")
	node.SetUuid(uuid)
	return node
}
func (self *Node) SetConn(conn net.Conn) {
}
func (n *Node) Write(b []byte) {
}

func (self *Node) LogicChannel() chan string {
	self.RLock()
	defer self.RUnlock()
	return self.logicChannel
}

func (self *Node) State() interface{} {
	self.RLock()
	defer self.RUnlock()
	return self.State_
}

func (self *Node) Set(key string, value interface{}) {
	self.Lock()
	defer self.Unlock()
	self.State_[key] = cast(value)
}

func (self *Node) Reset(key string) {
	switch self.State_[key].(type) {
	case int:
		self.Set(key, 0)
	case float64:
		self.Set(key, 0.0)
	case string:
		self.Set(key, "")
	case bool:
		self.Set(key, 0)
	}
}

func (self *Node) GetNotification() *notifications.Notification {
	return nil
}

func (self *Node) GetPing() bool {
	return false
}

func cast(s interface{}) interface{} {
	switch v := s.(type) {
	case int:
		return v
		//return strconv.Itoa(v)
	case float64:
		//return strconv.FormatFloat(v, 'f', -1, 64)
		return v
	case string:
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
		return v
	case bool:
		if v {
			return 1
		}
		return 0
	}
	return ""
}
