package main

import (
	"encoding/json"
	"io"
	"net"
	"syscall"

	log "github.com/cihub/seelog"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/logic"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/metrics"
	serverprotocol "github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/protocol"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/servernode"
)

type NodeServer struct {
	Config           *ServerConfig         `inject:""`
	Logic            *logic.Logic          `inject:""`
	Nodes            *serverprotocol.Nodes `inject:""`
	WebsocketHandler *WebsocketHandler     `inject:""`
	ElasticSearch    *ElasticSearch        `inject:""`
	Metrics          *metrics.Metrics      `inject:""`
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
	log.Info("New client connected")
	name := ""
	uuid := ""
	decoder := json.NewDecoder(connection)
	//encoder := json.NewEncoder(os.Stdout)
	var logicChannel chan string
	for {
		node := serverprotocol.NewNode()
		err := decoder.Decode(&node)

		if err != nil {
			//If the error was a network error we have disconnected. Otherwise it might be a json decode error
			if neterr, ok := err.(net.Error); (ok && !neterr.Temporary()) || err == io.EOF || err == syscall.ECONNRESET || err == syscall.EPIPE {
				log.Info(name, " - Client disconnected with error:", err.Error())
				if uuid != "" {
					ns.Nodes.Delete(uuid)
					close(logicChannel)
				}
				//TODO be able to not send everything always. perhaps implement remove instead of all?
				ns.WebsocketHandler.SendAllNodes()
				return
			}
			log.Warn("Not a net.Error but error: ", err)
			return
		}

		name = node.Name()
		uuid = node.Uuid()

		existingNode := ns.Nodes.ByUuid(uuid)
		if existingNode == nil {
			ns.Nodes.Add(node)
			logicChannel = ns.Logic.ListenForChanges(node.Uuid())
			node.SetConn(connection)
			ns.updateState(logicChannel, node)
		} else {
			existingNode.SetState(node.State())
			ns.updateState(logicChannel, existingNode)
		}

		ns.WebsocketHandler.SendSingleNode(uuid)
	}
}

func (ns *NodeServer) updateState(updateChan chan string, node serverprotocol.Node) {
	if node == nil {
		log.Warn("Recived an updateState but no node was provided, ignoring...")
		return
	}
	ns.Logic.Update(updateChan, node)
	ns.Metrics.Update(node)
}

func (self *NodeServer) addServerNode() {
	logicChannel := self.Logic.ListenForChanges(self.Config.Uuid)
	node := servernode.New(self.Config.Uuid, logicChannel)
	self.Nodes.Add(node)
}
