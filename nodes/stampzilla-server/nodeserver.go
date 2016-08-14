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
	"github.com/stampzilla/stampzilla-go/protocol"
	"github.com/stampzilla/stampzilla-go/protocol/devices"
)

type NodeServer struct {
	Config           *ServerConfig           `inject:""`
	Logic            *logic.Logic            `inject:""`
	Nodes            *serverprotocol.Nodes   `inject:""`
	Devices          *serverprotocol.Devices `inject:""`
	WsNodesHandler   *handlers.Nodes         `inject:""`
	WsDevicesHandler *handlers.Devices       `inject:""`
	ElasticSearch    *ElasticSearch          `inject:""`
	Metrics          *metrics.Metrics        `inject:""`
	Notifications    notifications.Router    `inject:""`
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
	ns.WsNodesHandler.SendDisconnectedNode(uuid)
	ns.Nodes.Delete(uuid)
	log.Info(name, " - Removing node from nodes list")

	devices := ns.Devices.SetOfflineByNode(uuid)
	for _, dev := range devices {
		ns.WsDevicesHandler.SendSingleDevice(dev)
	}

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

	updatePaket := protocol.NewUpdate()
	var logicChannel chan string
	for {

		err := decoder.Decode(&updatePaket)

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
				ns.WsNodesHandler.SendAllNodes()
				return
			}
			log.Warn("Not a net.Error but error: ", err)
			return
		}

		nodeIsAlive <- true

		existingNode := ns.Nodes.ByUuid(uuid)

		switch updatePaket.Type {
		case protocol.Pong:
			break
		case protocol.Ping:
			connection.Write([]byte("{\"Ping\":true}"))
		case protocol.UpdateNode:
			node := serverprotocol.NewNode()
			err := json.Unmarshal(*updatePaket.Data, &node)
			if err != nil {
				log.Errorf("%s", updatePaket.Data)
				log.Error(err)
				return
			}

			if existingNode != nil {
				log.Tracef("Existing node: %#v", existingNode)
				node.SetUuid(existingNode.Uuid()) // Add name and uuid to package
				node.SetName(existingNode.Name()) // Add name and uuid to package

				existingNode.SetState(node.State())
				existingNode.SetDevices(node.Devices())
				ns.updateState(logicChannel, existingNode)
				ns.syncDevices(existingNode.Devices(), node)

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
				ns.syncDevices(node.Devices(), node)
			}
			ns.WsNodesHandler.SendSingleNode(uuid)

		case protocol.Notification:
			if note := existingNode.GetNotification(*updatePaket.Data); note != nil {
				log.Tracef("Recived notification: %#v", note)
				ns.Notifications.Dispatch(*note) // Send the notification to the router
				continue
			}
		}

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

func (ns *NodeServer) syncDevices(newDevices devices.Map, node serverprotocol.Node) {
	for _, v := range newDevices {
		ns.Devices.Add(node.Uuid(), v)
		ns.WsDevicesHandler.SendSingleDevice(v)
	}
}

func (ns *NodeServer) addServerNode() {
	logicChannel := ns.Logic.ListenForChanges(ns.Config.Uuid)
	node := servernode.New(ns.Config.Uuid, logicChannel)
	err := ns.Nodes.Add(node)
	if err != nil {
		log.Critical(err)
		os.Exit(2)
	}
}
