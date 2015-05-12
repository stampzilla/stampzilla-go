package main

import (
	"encoding/json"
	"net"

	log "github.com/cihub/seelog"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/logic"
	serverprotocol "github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/protocol"
)

type NodeServer struct {
	Config           *ServerConfig         `inject:""`
	Logic            *logic.Logic          `inject:""`
	Nodes            *serverprotocol.Nodes `inject:""`
	WebsocketHandler *WebsocketHandler     `inject:""`
	ElasticSearch    *ElasticSearch        `inject:""`
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
	var logicChannel chan string
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
					close(logicChannel)
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

			if logicChannel == nil {
				logicChannel = ns.Logic.ListenForChanges(uuid)
			}

			existingNode := ns.Nodes.ByUuid(uuid)
			if existingNode == nil {
				ns.Nodes.Add(&node)
				//node.SetJsonEncoder(encoder)
				node.SetConn(connection)
			} else {
				existingNode.State = node.State
			}
			log.Info(node.Uuid, " - ", node.Name, " - Got update on state")
			ns.WebsocketHandler.SendSingleNode(uuid)

			//Send to logic for evaluation
			state, _ := json.Marshal(node.State)
			logicChannel <- string(state)

			// Try to send an update to elasticsearch
			if ns.ElasticSearch.StateUpdates != nil {
				select {
				case ns.ElasticSearch.StateUpdates <- &node: // Successfully deliverd to es
				default: // Failed to deliver to es
					log.Warn("Failed to update ElasticSearch")
				}
			}
		}

	}

}
