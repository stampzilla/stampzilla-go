package main

import (
	"encoding/json"
	"net"

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

	ServerNode *serverprotocol.Node
	State      *NodeState
}

func NewNodeServer() *NodeServer {
	return &NodeServer{}
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
	log.Info("New client connected")
	name := ""
	uuid := ""
	decoder := json.NewDecoder(connection)
	//encoder := json.NewEncoder(os.Stdout)
	for {
		var node serverprotocol.Node
		err := decoder.Decode(&node)

		if err != nil {
			if err.Error() == "EOF" {
				log.Info(name, " - Client disconnected")
				if uuid != "" {
					ns.Nodes.Delete(uuid)
					ns.Logic.StopListen(uuid)
				}
				//TODO be able to not send everything always. perhaps implement remove instead of all?
				ns.WebsocketHandler.SendAllNodes()
				return
			}
			log.Warn("Not disconnect but error: ", err)
			//return here?
		} else {
			name = node.Name
			uuid = node.Uuid

			existingNode := ns.Nodes.ByUuid(uuid)
			if existingNode == nil {
				ns.Nodes.Add(&node)
				//node.SetJsonEncoder(encoder)
				node.SetConn(connection)
				ns.updateState(&node)
			} else {
				existingNode.State = node.State
				ns.updateState(existingNode)
			}

			log.Info(node.Uuid, " - ", node.Name, " - Got update on state")
			ns.WebsocketHandler.SendSingleNode(uuid)
		}
	}
}

func (ns *NodeServer) updateState(node *serverprotocol.Node) {
	if node == nil {
		log.Warn("Recived an updateState but no node was provided, ignoring...")
		return
	}

	logicChannel := ns.Logic.ListenForChanges(node.Uuid)

	//Send to logic for evaluation
	state, _ := json.Marshal(node.State)
	*logicChannel <- string(state)

	//Send to metrics
	//TODO make this a buffered channel so we dont have to wait for the logging to complete before continueing.
	ns.Metrics.Update(node)
}

func (self *NodeServer) addServerNode() {
	node := &serverprotocol.Node{}
	node.Name = "server"

	self.State = &NodeState{
		values:     make(map[string]string),
		nodeServer: self,
	}
	self.State.logicChannel = self.Logic.ListenForChanges("server")
	node.SetState(self.State)

	self.Nodes.Add(node)
	self.ServerNode = node
}

func (self *NodeServer) Set(key, value string) {
	// Update the value
	self.State.values[key] = value
	self.updateState(self.ServerNode)
}

func (self *NodeServer) Trigger(key, value string) {
	// Update the value
	self.State.values[key] = value
	self.updateState(self.ServerNode)

	// Return the value to empty
	self.State.values[key] = ""
	self.updateState(self.ServerNode)
}

/* SERVER NODE STATE */
type NodeState struct {
	values       map[string]string
	logicChannel *chan string
	nodeServer   *NodeServer
}

func (self *NodeState) GetState() interface{} {
	return self.values
}
