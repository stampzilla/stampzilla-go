package main

import (
	"bufio"
	"encoding/json"
	"net"

	log "github.com/cihub/seelog"
	"github.com/stampzilla/stampzilla-go/stampzilla-server/logic"
	serverprotocol "github.com/stampzilla/stampzilla-go/stampzilla-server/protocol"
)

type NodeServer struct {
	Config           *ServerConfig         `inject:""`
	Logic            *logic.Logic          `inject:""`
	Nodes            *serverprotocol.Nodes `inject:""`
	WebsocketHandler *WebsocketHandler     `inject:""`
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
	for {
		reader := bufio.NewReader(connection)
		decoder := json.NewDecoder(reader)
		var info serverprotocol.Node
		err := decoder.Decode(&info)

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
			name = info.Name
			uuid = info.Uuid
			info.SetConn(connection)

			if logicChannel == nil {
				logicChannel = ns.Logic.ListenForChanges(uuid)
			}

			ns.Nodes.Add(&info)
			log.Info(info.Uuid, " - ", info.Name, " - Got update on state")
			ns.WebsocketHandler.SendSingleNode(uuid)

			//Send to logic for evaluation
			state, _ := json.Marshal(info.State)
			logicChannel <- string(state)
		}

	}

}
