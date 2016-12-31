package logic

import (
	"sync"
	"testing"

	serverprotocol "github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/protocol"
	"github.com/stretchr/testify/assert"
)

type nodeStub struct {
	serverprotocol.Node
	written [][]byte
	wg      *sync.WaitGroup
}

func (n *nodeStub) Name() string {
	return ""
}

func (n *nodeStub) Write(b []byte) (int, error) {
	n.written = append(n.written, b)
	if n.wg != nil {
		n.wg.Done()
	}
	return len(b), nil
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

	abort := make(chan struct{})
	command.Run(abort)

	assert.Equal(t, 1, len(nodes.node.written))
}
