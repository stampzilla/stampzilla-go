package protocol

import "testing"

func TestSearch(t *testing.T) {
	nodes := NewNodes()

	node := &Node{}
	node.Name = "Test"
	node.Uuid = "testuuid"

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
