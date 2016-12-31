package protocol

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddElement(t *testing.T) {

	n := NewNode("nodename")

	n.AddElement(&Element{
		Type:     ElementTypeText,
		Name:     "elname",
		Feedback: "asdf",
	})

	els := n.Elements()
	for _, v := range els {
		if v.Name == "elname" {
			assert.Equal(t, "asdf", v.Feedback)
			return
		}

	}

	t.Error("element not found")
}

func TestSetElement(t *testing.T) {

	n := NewNode("nodename")

	newEls := []*Element{
		&Element{
			Type:     ElementTypeText,
			Name:     "elname",
			Feedback: "asdf",
		},
	}

	n.SetElements(newEls)

	els := n.Elements()
	for _, v := range els {
		if v.Name == "elname" {
			assert.Equal(t, "asdf", v.Feedback)
			return
		}

	}

	t.Error("element not found")
}

func TestNodeGetters(t *testing.T) {

	n := NewNode("name")

	assert.Equal(t, "name", n.Name())
	n.SetUuid("uuid")
	n.SetName("newname")
	assert.Equal(t, "newname", n.Name())
	assert.Equal(t, "uuid", n.Uuid())

}

func TestNodeState(t *testing.T) {
	n := NewNode("name")
	n.SetState(123)
	assert.Equal(t, 123, n.State())
}

func TestNodeDevices(t *testing.T) {
	n := NewNode("name")
	assert.False(t, n.Devices().Exists("asdf"))
}
func TestNodeConfig(t *testing.T) {
	n := NewNode("name")
	n.SetUuid("uuid")
	n.Config().Add("asdf")
	if _, ok := n.Config().Config["uuid.asdf"]; !ok {
		t.Error("failed to find uuid.asdf in config")
	}
}
