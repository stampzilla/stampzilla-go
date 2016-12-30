package logic

import (
	log "github.com/cihub/seelog"
	serverprotocol "github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/protocol"
	"github.com/stampzilla/stampzilla-go/protocol"
)

type Command interface {
	Run(abort <-chan struct{})
}
type command struct {
	Command *protocol.Command `json:"command"`
	Uuid_   string            `json:"uuid"`
	nodes   serverprotocol.Searchable
}

func NewCommand(cmd *protocol.Command, uuid string) *command {
	return &command{Command: cmd, Uuid_: uuid}
}
func (c *command) Uuid() string {
	return c.Uuid_
}

//func (c *command) SetNodes(nodes serverprotocol.Searchable) {
//c.nodes = nodes
//}
func (c *command) Run(abort <-chan struct{}) {
	if c.nodes == nil {
		log.Warn("Node ", c.Uuid(), " - No nodes connected when tried to send: ", c.Command)
		return
	}
	node := c.nodes.Search(c.Uuid())
	if node != nil {
		msg := protocol.NewUpdateWithData(protocol.TypeCommand, &c.Command)

		err := node.WriteUpdate(msg)
		if err != nil {
			log.Warn("Node ", c.Uuid(), " - Failed to run command:", err)
			return
		}
		log.Infof("Running command %#v to %s", c.Command, c.Uuid())

		return
	}
	log.Warn("Node ", c.Uuid(), " not found :/,  lost command", c.Command)
}
