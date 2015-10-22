package logic

import (
	"testing"

	serverprotocol "github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/protocol"
	"github.com/stampzilla/stampzilla-go/protocol"
	"github.com/stretchr/testify/assert"
)

func TestActionsMapperSave(t *testing.T) {
	mapper := NewActionsMapper()

	actions := &Actions{}
	actions.Actions_ = []*action{
		&action{
			Commands: []*command{
				NewCommand(&protocol.Command{}, "uuid1"),
				NewCommand(&protocol.Command{}, "uuid2"),
			},
			Uuid_: "actionuuid1",
		},
		&action{
			Commands: []*command{
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
	a := &Actions{}
	a.Nodes = serverprotocol.NewNodes()
	node := serverprotocol.NewNode()
	node.SetName("nodename")
	node.SetUuid("nodeuuid")
	a.Nodes.Add(node)
	mapper.Load(a)

	assert.Equal(t, "nodeuuid", a.Nodes.ByName("nodename").Uuid())
	assert.NotNil(t, a.Nodes)
	assert.NotNil(t, a.Nodes.Search("nodeuuid"))

	assert.Equal(t, "actionuuid1", a.Actions_[0].Uuid())
	assert.Equal(t, "actionuuid2", a.Actions_[1].Uuid())

	assert.Equal(t, "actionuuid1", a.GetByUuid("actionuuid1").Uuid())
	assert.Equal(t, "actionuuid2", a.GetByUuid("actionuuid2").Uuid())

	assert.Nil(t, a.GetByUuid("unknown"))

	assert.Equal(t, "uuid1", a.Actions_[1].Commands[0].Uuid())
	assert.Equal(t, "uuid2", a.Actions_[1].Commands[1].Uuid())

	//assert we set nodes dependency correctly on Command
	assert.Equal(t, "nodename", a.Actions_[1].Commands[0].nodes.Search("nodeuuid").Name())

	//fmt.Printf("%#v\n", a)
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
