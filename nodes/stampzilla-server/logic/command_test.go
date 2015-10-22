package logic

import (
	"testing"

	serverprotocol "github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/protocol"
	"github.com/stretchr/testify/assert"
)

type nodeStub struct {
	serverprotocol.Node
	written [][]byte
}

func (n *nodeStub) Name() string {
	return ""
}

func (n *nodeStub) Write(b []byte) {
	n.written = append(n.written, b)
}

type nodesStub struct {
	node *nodeStub
}

func (n *nodesStub) Search(what string) serverprotocol.Node {
	return n.node
}

func TestRunCommand(t *testing.T) {
	nodes := &nodesStub{}
	nodes.node = &nodeStub{}

	command := &command{}
	command.Uuid_ = "cmduuid"
	command.nodes = nodes

	command.Run()

	assert.Equal(t, 1, len(nodes.node.written))
}
