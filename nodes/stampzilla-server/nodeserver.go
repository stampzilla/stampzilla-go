package main

import (
	"encoding/json"
	"net"
	"strconv"

	log "github.com/cihub/seelog"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/logic"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/metrics"
	serverprotocol "github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/protocol"
)

type NodeServer struct {
	Config           *ServerConfig         `inject:""`
	Logic            *logic.Logic          `inject:""`
	Nodes            *serverprotocol.Nodes `inject:""`
	WebsocketHandler *WebsocketHandler     `inject:""`
	ElasticSearch    *ElasticSearch        `inject:""`
	Metrics          *metrics.Metrics      `inject:""`

	ServerNode   *serverprotocol.Node
	State        map[string]interface{}
	logicChannel chan string
}

func NewNodeServer() *NodeServer {
	return &NodeServer{
		ServerNode: &serverprotocol.Node{},
		State:      make(map[string]interface{}),
	}
}

func (ns *NodeServer) Start() {
	log.Info("Starting NodeServer (:" + ns.Config.NodePort + ")")
	listen, err := net.Listen("tcp", ":"+ns.Config.NodePort)
	if err != nil {
		log.Error("listen error", err)
		return
	}

	ns.Logic.RestoreRulesFromFile("rules.json")
	ns.addServerNode()

	ns.logicChannel = ns.Logic.ListenForChanges(ns.ServerNode.Uuid)

	//ns.Logic.Listen()

	//return
	go func() {
		for {
			fd, err := listen.Accept()
			if err != nil {
				log.Error("accept error", err)
				return
			}

			go ns.newNodeConnection(fd)
		}
	}()
}

func (ns *NodeServer) newNodeConnection(connection net.Conn) {
	// Recive data
	log.Info("New client connected (", connection.RemoteAddr(), ")")
	name := ""
	uuid := ""
	decoder := json.NewDecoder(connection)
	//encoder := json.NewEncoder(os.Stdout)
	var logicChannel chan string
	for {
		var node serverprotocol.Node
		err := decoder.Decode(&node)

		if err != nil {
			if err.Error() == "EOF" {
				log.Info(name, " - Client disconnected")
				if uuid != "" {
					ns.Nodes.Delete(uuid)
					close(logicChannel)
				}
				//TODO be able to not send everything always. perhaps implement remove instead of all?
				ns.WebsocketHandler.SendAllNodes()
				return
			}
			log.Warn(name, " - Not disconnect but error (closing connection): ", err)

			connection.Close()
			return
		} else {
			name = node.Name
			uuid = node.Uuid

			existingNode := ns.Nodes.ByUuid(uuid)
			if existingNode == nil {
				ns.Nodes.Add(&node)
				logicChannel = ns.Logic.ListenForChanges(node.Uuid)
				node.SetConn(connection)
				ns.updateState(logicChannel, &node)
			} else {
				existingNode.SetState(node.State())
				ns.updateState(logicChannel, existingNode)
			}

			ns.WebsocketHandler.SendSingleNode(uuid)
		}
	}
}

func (ns *NodeServer) updateState(updateChan chan string, node *serverprotocol.Node) {
	if node == nil {
		log.Warn("Recived an updateState but no node was provided, ignoring...")
		return
	}

	//log.Info(node.Uuid, " - ", node.Name, " - Got update on state")

	//Send to logic for evaluation
	//logicChannel := ns.Logic.ListenForChanges(node.Uuid)
	//state, _ := json.Marshal(node.State)
	//*logicChannel <- string(state)
	ns.Logic.Update(updateChan, node)

	//Send to metrics
	ns.Metrics.Update(node)
}

func (self *NodeServer) addServerNode() {
	self.ServerNode.Name = "server"
	self.ServerNode.Uuid = self.Config.Uuid
	self.ServerNode.SetState(self.State)

	self.Nodes.Add(self.ServerNode)
}

func (self *NodeServer) Set(key string, value interface{}) {
	self.State[key] = cast(value)
	self.updateState(self.logicChannel, self.ServerNode)
}

func (self *NodeServer) Trigger(key string, value interface{}) {
	self.Set(key, value)

	switch self.State[key].(type) {
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
