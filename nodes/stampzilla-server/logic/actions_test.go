package logic

import (
	"testing"

	serverprotocol "github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/protocol"
	"github.com/stampzilla/stampzilla-go/protocol"
	"github.com/stretchr/testify/assert"
)

func TestActionsMapperSave(t *testing.T) {
	mapper := newActionsMapper()

	actions := &actions{}
	actions.Actions = []*action{
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

	mapper := newActionsMapper()
	a := &actions{}
	a.nodes = serverprotocol.NewNodes()
	node := serverprotocol.NewNode()
	node.SetName("nodename")
	node.SetUuid("nodeuuid")
	a.nodes.Add(node)
	mapper.Load(a)

	assert.Equal(t, "nodeuuid", a.nodes.ByName("nodename").Uuid())
	assert.NotNil(t, a.nodes)
	assert.NotNil(t, a.nodes.ByUuid("nodeuuid"))

	assert.Equal(t, "actionuuid1", a.Actions[0].Uuid())
	assert.Equal(t, "actionuuid2", a.Actions[1].Uuid())

	assert.Equal(t, "uuid1", a.Actions[1].Commands[0].Uuid())
	assert.Equal(t, "uuid2", a.Actions[1].Commands[1].Uuid())

	//fmt.Printf("%#v\n", a)
}
