package logic

import (
	"flag"
	"os"
	"sync"
	"testing"
	"time"

	log "github.com/cihub/seelog"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	flag.Parse()

	consoleWriter, _ := log.NewConsoleWriter()
	formatter, _ := log.NewFormatter("[%Level] %File:%Line - %Msg%n")
	root, _ := log.NewSplitDispatcher(formatter, []interface{}{consoleWriter})

	constraints, _ := log.NewMinMaxConstraints(log.ErrorLvl, log.CriticalLvl)
	if testing.Verbose() {
		constraints, _ = log.NewMinMaxConstraints(log.DebugLvl, log.CriticalLvl)
	}

	ex, _ := log.NewLogLevelException("*", "*main.go", constraints)
	exceptions := []*log.LogLevelException{ex}

	logger := log.NewAsyncLoopLogger(log.NewLoggerConfig(constraints, exceptions, root))
	log.ReplaceLogger(logger)

	os.Exit(m.Run())
}

func TestRunAction(t *testing.T) {
	nodes := &nodesStub{}
	nodes.node = &nodeStub{}
	nodes.node.wg = &sync.WaitGroup{}

	cmd := &command{}
	action := &action{}
	action.Commands = append(action.Commands, cmd)
	action.Commands = append(action.Commands, cmd)
	action.Commands = append(action.Commands, cmd)
	action.Commands = append(action.Commands, cmd)
	action.Commands = append(action.Commands, cmd)

	nodes.node.wg.Add(5)
	action.Run(nil)

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
	action.Run(nil)

	nodes.node.wg.Wait() // Wait on all to finish
	assert.Equal(t, 4, len(nodes.node.written), "Not all commands did run")

	nodes.node.wg.Add(4)
	pause.wg.Add(3)
	action.Run(nil)
	<-time.After(time.Millisecond * 30) // Cancel in the middle of the second pause
	action.Cancel()
	<-time.After(time.Millisecond * 150) // Make sure to wait on everything to finish

	assert.Equal(t, 6, len(nodes.node.written), "Action was NOT canceled")
}

type pauseStub struct {
	commandPause

	wg *sync.WaitGroup
}

func (p *pauseStub) Run(a <-chan struct{}) {
	abort := make(chan struct{})
	p.commandPause.Run(abort)
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
	action.Run(nil)

	cmd.wg.Wait()
	t1 := time.Now()

	assert.WithinDuration(t, t0.Add(time.Millisecond*500), t1, time.Millisecond*10)
}
