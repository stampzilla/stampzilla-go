package logic

import (
	"encoding/json"
	"fmt"

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
	log.Info("Running command", ra.Command,"to",ra.Uuid)
	if ra.nodes == nil {
		fmt.Println("ra.nodes is nil!")
		return
	}
	node := ra.nodes.Search(ra.Uuid)
	if node != nil {
		jsonToSend, err := json.Marshal(&ra.Command)
		if err != nil {
			log.Error(err)
			return
		}
		node.Conn().Write(jsonToSend)
	}
}
