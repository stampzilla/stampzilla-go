package main

import (
	"encoding/json"
	"io"
	"net"
	"os"
	"syscall"
	"time"

	log "github.com/cihub/seelog"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/logic"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/metrics"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/notifications"
	serverprotocol "github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/protocol"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/servernode"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/websocket/handlers"
)

type NodeServer struct {
	Config           *ServerConfig         `inject:""`
	Logic            *logic.Logic          `inject:""`
	Nodes            *serverprotocol.Nodes `inject:""`
	WebsocketHandler *handlers.Nodes       `inject:""`
	ElasticSearch    *ElasticSearch        `inject:""`
	Metrics          *metrics.Metrics      `inject:""`
	Notifications    notifications.Router  `inject:""`
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

func (ns *NodeServer) NodeDisconnected(uuid, name string) {
	ns.WebsocketHandler.SendDisconnectedNode(uuid)
	ns.Nodes.Delete(uuid)
	log.Info(name, " - Removing node from nodes list")

	// Span a goroutine to check if node is disconnected after a delay
	go func() {
		<-time.After(time.Minute)

		if ns.Nodes.ByUuid(uuid) != nil {
			// Everything is great, node connected igain
			return
		}

		// Bad, node still not connected
		notify.Warn("Node disconnected -> " + name + "(" + uuid + ")")
	}()
}

func (ns *NodeServer) newNodeConnection(connection net.Conn) {
	// Recive data
	log.Trace("New client connected (", connection.RemoteAddr(), ")")
	name := ""
	uuid := ""
	decoder := json.NewDecoder(connection)
	//encoder := json.NewEncoder(os.Stdout)

	nodeIsAlive := make(chan bool)
	go timeoutMonitor(connection, nodeIsAlive)

	var logicChannel chan string
	for {
		node := serverprotocol.NewNode()
		err := decoder.Decode(&node)

		if err != nil {
			//If the error was a network error we have disconnected. Otherwise it might be a json decode error
			if neterr, ok := err.(net.Error); (ok && !neterr.Temporary()) || err == io.EOF || err == syscall.ECONNRESET || err == syscall.EPIPE {
				log.Info(name, " - Client disconnected with error:", err.Error())
				connection.Close()

				if uuid != "" {
					close(logicChannel)
					ns.NodeDisconnected(uuid, name)
					return
				}
				// No uuid available, send the whole node list to webclients
				ns.WebsocketHandler.SendAllNodes()
				return
			}
			log.Warn("Not a net.Error but error: ", err)
			return
		}

		nodeIsAlive <- true

		if node.GetPong() {
			continue
		}
		if node.GetPing() {
			connection.Write([]byte("{\"Ping\":true}"))
			continue
		}

		if existingNode := ns.Nodes.ByUuid(uuid); existingNode != nil {
			log.Tracef("Existing node: %#v", existingNode)
			node.SetUuid(existingNode.Uuid()) // Add name and uuid to package
			node.SetName(existingNode.Name()) // Add name and uuid to package

			if note := node.GetNotification(); note != nil {
				log.Tracef("Recived notification: %#v", note)
				ns.Notifications.Dispatch(*note) // Send the notification to the router
				continue
			}

			existingNode.SetState(node.State())
			ns.updateState(logicChannel, existingNode)

			existingNode.SetElements(node.Elements())
		} else {
			name = node.Name()
			uuid = node.Uuid()

			err := ns.Nodes.Add(node)
			if err != nil {
				log.Error(err)
				continue
			}

			log.Info("New client connected (", name, " - ", uuid, ")")

			logicChannel = ns.Logic.ListenForChanges(node.Uuid())
			node.SetConn(connection)
			ns.updateState(logicChannel, node)
		}

		ns.WebsocketHandler.SendSingleNode(uuid)
	}
}

func timeoutMonitor(connection net.Conn, nodeIsAlive chan bool) {
	log.Debug("Timeout monitor started (", connection.RemoteAddr(), ")")
	defer log.Debug("Timeout monitor closed (", connection.RemoteAddr(), ")")

	for {
		select {
		case <-nodeIsAlive:
			// Everything is good, just continue
			continue
		case <-time.After(time.Second * 10):
			// Send ping and wait for the answer
			connection.Write([]byte("{\"Ping\":true}"))

			select {
			case <-nodeIsAlive:
				continue
			case <-time.After(time.Second * 2):
				log.Warn("Connection timeout, no answer on ping")
				connection.Close()
				return
			}
		}
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
	err := self.Nodes.Add(node)
	if err != nil {
		log.Critical(err)
		os.Exit(2)
	}
}
