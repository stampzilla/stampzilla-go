package logic

import (
	"log"
	"sync"
	"testing"

	serverprotocol "github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/protocol"
	"github.com/stampzilla/stampzilla-go/protocol"
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

func (n *nodeStub) Write(b []byte) error {
	n.written = append(n.written, b)
	if n.wg != nil {
		n.wg.Done()
	}
	return nil
}

func (n *nodeStub) WriteUpdate(msg *protocol.Update) error {
	bytes, err := msg.ToJSON()
	if err != nil {
		log.Println(err)
		return err
	}

	return n.Write(bytes)
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
