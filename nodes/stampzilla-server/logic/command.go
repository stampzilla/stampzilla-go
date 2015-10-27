package logic

import (
	"encoding/json"

	log "github.com/cihub/seelog"
	serverprotocol "github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/protocol"
	"github.com/stampzilla/stampzilla-go/protocol"
)

type Command interface {
	Run()
	Uuid() string
	SetNodes(serverprotocol.Searchable)
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
func (c *command) SetNodes(nodes serverprotocol.Searchable) {
	c.nodes = nodes
}
func (c *command) Run() {
	if c.nodes == nil {
		log.Warn("Node ", c.Uuid(), " - No nodes connected when tried to send: ", c.Command)
		return
	}
	node := c.nodes.Search(c.Uuid())
	if node != nil {
		jsonToSend, err := json.Marshal(&c.Command)
		if err != nil {
			log.Warn("Node ", c.Uuid(), " - Failed to marshal command: ", c.Command)
			log.Error(err)
			return
		}

		log.Info("Running command ", c.Command, " to ", c.Uuid())
		node.Write(jsonToSend)

		return
	}
	log.Warn("Node ", c.Uuid(), " not found :/,  lost command", c.Command)
}
