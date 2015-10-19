package logic

import (
	"testing"
	"time"

	serverprotocol "github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/protocol"
)

type nodeStub struct {
	serverprotocol.Node
}

func (n *nodeStub) Name() string {
	return ""
}

func (n *nodeStub) Write(b []byte) {

}

type nodesStub struct {
}

func (n *nodesStub) Search(what string) serverprotocol.Node {
	return &nodeStub{}
	return nil
}

func TestRunCommand(t *testing.T) {

	nodes := &nodesStub{}

	command := &command{}
	command.Uuid_ = "cmduuid"
	command.nodes = nodes

	command.Run()

	time.Sleep(time.Second)
	//fmt.Printf("%#v\n", a)
}
