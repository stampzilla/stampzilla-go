package logic

import (
	"encoding/json"
	"fmt"

	log "github.com/cihub/seelog"
	"github.com/stampzilla/stampzilla-go/protocol"
)

type RuleAction interface {
	RunCommand()
}
type ruleAction struct {
	Command *protocol.Command `json:"command"`
	Uuid    string            `json:"uuid"`
}

func NewRuleAction(cmd *protocol.Command, uuid string) RuleAction {
	return &ruleAction{cmd, uuid}
}
func (ra *ruleAction) RunCommand() {
	fmt.Println("Running command", ra.Command)
	if nodes == nil {
		return
	}
	node := nodes.Search(ra.Uuid)
	if node != nil {
		jsonToSend, err := json.Marshal(&ra.Command)
		if err != nil {
			log.Error(err)
			return
		}
		node.Conn().Write(jsonToSend)
	}
}
