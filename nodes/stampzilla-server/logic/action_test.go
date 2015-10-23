package logic

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunAction(t *testing.T) {
	nodes := &nodesStub{}
	nodes.node = &nodeStub{}
	nodes.node.wg = &sync.WaitGroup{}

	cmd := &command{}
	cmd.Uuid_ = "cmduuid"
	cmd.nodes = nodes
	action := &action{}
	action.Commands = append(action.Commands, cmd)
	action.Commands = append(action.Commands, cmd)
	action.Commands = append(action.Commands, cmd)
	action.Commands = append(action.Commands, cmd)
	action.Commands = append(action.Commands, cmd)

	nodes.node.wg.Add(5)
	action.Run()

	nodes.node.wg.Wait()
	assert.Equal(t, 5, len(nodes.node.written))
}
func TestCancelAction(t *testing.T) {
	nodes := &nodesStub{}
	nodes.node = &nodeStub{}
	nodes.node.wg = &sync.WaitGroup{}

	cmd := &command{}
	cmd.Uuid_ = "cmduuid"
	cmd.nodes = nodes
	action := &action{}
	action.Commands = append(action.Commands, cmd)
	action.Commands = append(action.Commands, cmd)
	action.Commands = append(action.Commands, cmd)

	nodes.node.wg.Add(3)
	action.Run()

	nodes.node.wg.Wait()
	assert.Equal(t, 3, len(nodes.node.written))

	action.Run()
	action.Cancel()
	assert.Equal(t, 3, len(nodes.node.written))
}
