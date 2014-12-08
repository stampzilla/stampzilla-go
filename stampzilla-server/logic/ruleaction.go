package logic

import (
	"encoding/json"

	log "github.com/cihub/seelog"
	"github.com/stampzilla/stampzilla-go/protocol"
	serverprotocol "github.com/stampzilla/stampzilla-go/stampzilla-server/protocol"
)

type RuleAction interface {
	RunCommand()
}
type ruleAction struct {
	Command *protocol.Command `json:"command"`
	Uuid    string            `json:"uuid"`
	nodes   *serverprotocol.Nodes
}

func NewRuleAction(cmd *protocol.Command, uuid string) RuleAction {
	return &ruleAction{Command: cmd, Uuid: uuid}
}
func (ra *ruleAction) RunCommand() {
	if ra.nodes == nil {
		log.Warn("Node ", ra.Uuid, " - No nodes connected when tried to send: ", ra.Command)
		return
	}
	node := ra.nodes.Search(ra.Uuid)
	if node != nil {
		jsonToSend, err := json.Marshal(&ra.Command)
		if err != nil {
			log.Warn("Node ", ra.Uuid, " - Failed to marshal command: ", ra.Command)
			log.Error(err)
			return
		}

		log.Info("Running command ", ra.Command, " to ", ra.Uuid)
		node.Conn().Write(jsonToSend)
		return
	}
	log.Warn("Node ", ra.Uuid, " not found :/,  lost command", ra.Command)
}
