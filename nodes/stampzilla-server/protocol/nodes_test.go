package protocol

import (
	"bufio"
	"net"
	"testing"
)

func TestSearch(t *testing.T) {
	nodes := NewNodes()

	node := &node{}
	node.SetName("Test")
	node.SetUuid("testuuid")

	nodes.Add(node)

	found := nodes.Search("Test")
	if found == nil {
		t.Error("nodes.Search expected: node with Name test got nil")
	}

	found = nodes.Search("testuuid")
	if found == nil {
		t.Error("nodes.Search expected: node with Uuid testuuid got nil")
	}

	found = nodes.Search("notfound")
	if found != nil {
		t.Errorf("nodes.Search expected: nil got: %s", found.Name)
	}

}
func TestDelete(t *testing.T) {
	nodes := NewNodes()
	node := &node{}
	node.SetName("Test")
	node.SetUuid("testuuid")
	nodes.Add(node)

	nodes.Delete("testuuid")

	if len(nodes.All()) != 0 {
		t.Error("expected nodes length to be 0")
	}

}

func TestWrite(t *testing.T) {
	cli, srv := net.Pipe()
	defer cli.Close()

	node := &node{}
	node.SetName("Test")
	node.SetUuid("testuuid")
	node.SetConn(cli)

	go func() {
		node.Write([]byte("hej"))
	}()
	status, err := bufio.NewReader(srv).ReadString('\n')
	if err != nil {
		t.Error(err)
		return
	}
	if status != "hej\n" {
		t.Error("Expected to have read \"hej\" got: ", status)
	}
}
