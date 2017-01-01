package logic

import (
	"testing"

	serverprotocol "github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/protocol"
	"github.com/stampzilla/stampzilla-go/protocol"
	"github.com/stretchr/testify/assert"
)

func TestActionsMapperSave(t *testing.T) {
	mapper := NewActionsMapper()

	actions := &ActionService{}
	actions.actions = []Action{
		&action{
			Commands: []Command{
				NewCommand(&protocol.Command{}, "uuid1"),
				NewPause("10ms"),
				NewCommand(&protocol.Command{}, "uuid2"),
				&command_notify{},
			},
			Uuid_: "actionuuid1",
		},
		&action{
			Commands: []Command{
				NewCommand(&protocol.Command{}, "uuid1"),
				NewCommand(&protocol.Command{}, "uuid2"),
			},
			Uuid_: "actionuuid2",
		},
	}
	mapper.Save(actions)
}
func TestActionsMapperLoad(t *testing.T) {

	mapper := NewActionsMapper()
	a := &ActionService{}
	a.Nodes = serverprotocol.NewNodes()
	node := serverprotocol.NewNode()
	node.SetName("nodename")
	node.SetUuid("nodeuuid")
	a.Nodes.Add(node)
	mapper.Load(a)

	assert.Equal(t, "nodeuuid", a.Nodes.ByName("nodename").Uuid())
	assert.NotNil(t, a.Nodes)
	assert.NotNil(t, a.Nodes.Search("nodeuuid"))

	assert.Equal(t, "actionuuid1", a.actions[0].Uuid())
	assert.Equal(t, "actionuuid2", a.actions[1].Uuid())

	assert.Equal(t, "actionuuid1", a.GetByUuid("actionuuid1").Uuid())
	assert.Equal(t, "actionuuid2", a.GetByUuid("actionuuid2").Uuid())

	assert.Nil(t, a.GetByUuid("unknown"))

	assert.IsType(t, &command{}, a.actions[0].(*action).Commands[0])
	assert.Equal(t, "uuid1", a.actions[0].(*action).Commands[0].(*command).Uuid_)

	assert.IsType(t, &command_pause{}, a.actions[0].(*action).Commands[1])

	assert.IsType(t, &command{}, a.actions[0].(*action).Commands[2])
	assert.Equal(t, "uuid2", a.actions[0].(*action).Commands[2].(*command).Uuid_)

	if cmd, ok := a.actions[1].(*action).Commands[0].(*command); ok {
		assert.Equal(t, "nodename", cmd.nodes.Search("nodeuuid").Name())
	} else {
		t.Error("Wrong type for command, should be *command")
	}

	assert.IsType(t, &command_notify{}, a.actions[0].(*action).Commands[3])

	//fmt.Printf("%#v\n", a)
}

func TestActionServiceRun(t *testing.T) {
	a := &ActionService{}
	a.Start()

	assert.Len(t, a.Get(), 2)
}

//func TestActionsRun(t *testing.T) {

//mapper := NewActionsMapper()
//a := &actions{}
//a.nodes = serverprotocol.NewNodes()
//node := serverprotocol.NewNode()
//node.SetName("nodename")
//node.SetUuid("nodeuuid")
//a.nodes.Add(node)
//mapper.Load(a)

//a.Run()

//TODO assert things here!

//fmt.Printf("%#v\n", a)
//}
