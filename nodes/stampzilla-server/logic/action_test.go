package logic

import (
	"sync"
	"testing"
	"time"

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

	pause := &pauseStub{}
	pause.SetDuration("20ms")
	pause.wg = &sync.WaitGroup{}

	cmd := &command{}
	cmd.Uuid_ = "cmduuid"
	cmd.nodes = nodes

	action := &action{}
	action.Commands = append(action.Commands, cmd) // 0ms
	action.Commands = append(action.Commands, pause)
	action.Commands = append(action.Commands, cmd) // 20ms
	action.Commands = append(action.Commands, pause)
	action.Commands = append(action.Commands, cmd) // 40ms
	action.Commands = append(action.Commands, pause)
	action.Commands = append(action.Commands, cmd) // 60ms

	nodes.node.wg.Add(4)
	pause.wg.Add(3)
	action.Run()
	pause.wg.Wait()

	assert.Equal(t, 4, len(nodes.node.written), "Not all commands did run")

	nodes.node.wg.Add(4)
	pause.wg.Add(3)
	action.Run()
	<-time.After(time.Millisecond * 30)
	action.Cancel()
	<-time.After(time.Millisecond * 100)

	assert.Equal(t, 6, len(nodes.node.written), "Action was NOT canceled")
}

type pauseStub struct {
	command_pause

	wg *sync.WaitGroup
}

func (p *pauseStub) Run() {
	p.command_pause.Run()
	p.wg.Done()
}

func TestPauseAction(t *testing.T) {
	cmd := &pauseStub{}
	cmd.SetDuration("100ms")
	cmd.wg = &sync.WaitGroup{}
	cmd.wg.Add(5)

	action := &action{}
	action.Commands = append(action.Commands, cmd)
	action.Commands = append(action.Commands, cmd)
	action.Commands = append(action.Commands, cmd)
	action.Commands = append(action.Commands, cmd)
	action.Commands = append(action.Commands, cmd)

	t0 := time.Now()
	action.Run()

	cmd.wg.Wait()
	t1 := time.Now()

	assert.WithinDuration(t, t0.Add(time.Millisecond*500), t1, time.Millisecond*10)
}
