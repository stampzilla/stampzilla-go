package main

import (
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gozwave/gozw/application"
	"github.com/stampzilla/stampzilla-go/nodes/basenode"
	"github.com/stampzilla/stampzilla-go/pkg/notifier"
	"github.com/stampzilla/stampzilla-go/protocol"
)

var VERSION string = "dev"
var BUILD_DATE string = ""

type node struct {
	notify     *notifier.Notify
	node       *protocol.Node
	connection basenode.Connection
	state      *state
	controller *application.Layer
}

func (n *node) Start() {
	config := basenode.NewConfig()
	basenode.SetConfig(config)

	n.node = protocol.NewNode("gozw")
	n.node.Version = VERSION
	n.node.BuildDate = BUILD_DATE
	n.node.SetUuid(config.Uuid)

	n.state = newState()
	n.node.SetState(n.state)

	//Start communication with the server
	n.connection = basenode.Connect()
	n.notify = notifier.New(n.connection)
	n.notify.SetSource(n.node)

	go n.monitorState()
	go n.serverRecv()

	go n.debounceStateUpdates(time.Millisecond*50, n.state.updateNotifications)

	n.node.AddElement(&protocol.Element{
		Type: protocol.ElementTypeButton,
		Name: "Include node",
		Command: &protocol.Command{
			Cmd: "add",
		},
	})
	n.node.AddElement(&protocol.Element{
		Type: protocol.ElementTypeButton,
		Name: "Exclude node",
		Command: &protocol.Command{
			Cmd: "remove",
		},
	})
}

func (n *node) monitorState() {
	for s := range n.connection.State() {
		switch s {
		case basenode.ConnectionStateConnected:
			n.state.RLock()
			n.connection.Send(n.node.Node())
			n.state.RUnlock()
		case basenode.ConnectionStateDisconnected:
		}
	}
}

func (n *node) serverRecv() {
	for d := range n.connection.Receive() {
		n.processCommand(d)
	}
}

func (n *node) processCommand(cmd protocol.Command) {
	switch cmd.Cmd {
	case "add":
		node, err := n.controller.AddNode()
		spew.Dump(node)
		spew.Dump(err)
	case "remove":
		node, err := n.controller.RemoveNode()
		spew.Dump(node)
		spew.Dump(err)
	}
}

func (n *node) debounceStateUpdates(interval time.Duration, input chan struct{}) {
	for {
		c := time.After(interval)

		select {
		case <-input:
		case <-c:
			n.state.RLock()
			n.connection.Send(n.node.Node())
			n.state.RUnlock()
			<-input
			n.state.RLock()
			n.connection.Send(n.node.Node())
			n.state.RUnlock()
		}
	}
}
