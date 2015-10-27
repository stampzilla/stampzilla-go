package logic

import (
	"encoding/json"
	"fmt"
	"log"
	"testing"
	"time"

	serverprotocol "github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/protocol"
	"github.com/stretchr/testify/assert"
)

type ruleActionStub struct {
	actionCount *int
}

func (ra *ruleActionStub) Run() {
	fmt.Println("RuleActionStubRAN")
	*ra.actionCount++
}
func (ra *ruleActionStub) Cancel() {
}
func (ra *ruleActionStub) Uuid() string {
	return ""
}
func (ra *ruleActionStub) Name() string {
	return ""
}

func NewRuleActionStub(actionCount *int) *ruleActionStub {
	return &ruleActionStub{actionCount}
}

func TestParseRuleEnterExitActionsEvaluateTrue(t *testing.T) {

	logic := NewLogic()

	rule := logic.AddRule("test rule 1")

	actionRunCount := 0
	action := NewRuleActionStub(&actionRunCount)
	rule.AddEnterAction(action)
	rule.AddExitAction(action)

	rule.AddCondition(&ruleCondition{`Devices[1].State`, "==", true, "uuid1234"})
	rule.AddCondition(&ruleCondition{`Devices[2].State`, "!=", "OFF", "uuid1234"})
	//rule.AddCondition(&ruleCondition{`Devices[3].State`, "!=", "OFF"})

	state := `
		{
			"Devices": {
				"1": {
					"Id": "1",
					"Name": "Dev1",
					"State": true,
					"Type": ""
				},
				"2": {
					"Id": "2",
					"Name": "Dev2",
					"State": "ON",
					"Type": ""
				}
			}
		}
	`

	logic.SetState("uuid1234", state)
	logic.EvaluateRules()

	state = `
		{
			"Devices": {
				"1": {
					"Id": "1",
					"Name": "Dev1",
					"State": "OFF",
					"Type": ""
				},
				"2": {
					"Id": "2",
					"Name": "Dev2",
					"State": "OFF",
					"Type": ""
				}
			}
		}
	`
	logic.SetState("uuid1234", state)
	logic.EvaluateRules()

	if len(logic.States()) != 1 {
		t.Errorf("length of logic.States should be 1. got: %s", len(logic.States()))
	}

	fmt.Println(actionRunCount)
	if actionRunCount == 2 {
		return
	}
	t.Errorf("actionRunCount wrong expected: %s got %s", 2, actionRunCount)
}

func TestParseRuleEnterExitActionsEvaluateFalse(t *testing.T) {

	logic := NewLogic()

	rule := logic.AddRule("test rule 1")

	actionRunCount := 0
	action := NewRuleActionStub(&actionRunCount)
	rule.AddEnterAction(action)
	rule.AddExitAction(action)

	rule.AddCondition(&ruleCondition{`Devices[1].State`, "==", false, "uuid1234"})
	rule.AddCondition(&ruleCondition{`Devices[2].State`, "!=", "OFF", "uuid1234"})
	//rule.AddCondition(&ruleCondition{`Devices[3].State`, "!=", "OFF"})

	state := `
		{
			"Devices": {
				"1": {
					"Id": "1",
					"Name": "Dev1",
					"State": true,
					"Type": ""
				},
				"2": {
					"Id": "2",
					"Name": "Dev2",
					"State": "ON",
					"Type": ""
				}
			}
		}
	`

	logic.SetState("uuid1234", state)
	logic.EvaluateRules()

	state = `
		{
			"Devices": {
				"1": {
					"Id": "1",
					"Name": "Dev1",
					"State": "OFF",
					"Type": ""
				},
				"2": {
					"Id": "2",
					"Name": "Dev2",
					"State": "OFF",
					"Type": ""
				}
			}
		}
	`
	logic.SetState("uuid1234", state)
	logic.EvaluateRules()

	if len(logic.States()) != 1 {
		t.Errorf("length of logic.States should be 1. got: %s", len(logic.States()))
	}

	fmt.Println(actionRunCount)
	if actionRunCount == 0 {
		return
	}
	t.Errorf("actionRunCount wrong expected: %s got %s", 0, actionRunCount)
}

func TestParseRuleEnterExitActionsWithoutUuid(t *testing.T) {

	logic := NewLogic()

	rule := logic.AddRule("test rule 1")

	actionRunCount := 0
	action := NewRuleActionStub(&actionRunCount)
	rule.AddEnterAction(action)
	rule.AddExitAction(action)

	rule.AddCondition(&ruleCondition{`Devices[1].State`, "==", false, ""})
	rule.AddCondition(&ruleCondition{`Devices[2].State`, "!=", "OFF", ""})
	//rule.AddCondition(&ruleCondition{`Devices[3].State`, "!=", "OFF"})

	state := `
		{
			"Devices": {
				"1": {
					"Id": "1",
					"Name": "Dev1",
					"State": true,
					"Type": ""
				},
				"2": {
					"Id": "2",
					"Name": "Dev2",
					"State": "ON",
					"Type": ""
				}
			}
		}
	`

	logic.SetState("uuid1234", state)
	logic.EvaluateRules()

	state = `
		{
			"Devices": {
				"1": {
					"Id": "1",
					"Name": "Dev1",
					"State": "OFF",
					"Type": ""
				},
				"2": {
					"Id": "2",
					"Name": "Dev2",
					"State": "OFF",
					"Type": ""
				}
			}
		}
	`
	logic.SetState("uuid1234", state)
	logic.EvaluateRules()

	if len(logic.States()) != 1 {
		t.Errorf("length of logic.States should be 1. got: %s", len(logic.States()))
	}

	fmt.Println(actionRunCount)
	if actionRunCount == 0 {
		return
	}
	t.Errorf("actionRunCount wrong expected: %s got %s", 0, actionRunCount)
}

func TestListenForChanges(t *testing.T) {

	logic := NewLogic()

	rule := logic.AddRule("test rule 1")

	actionRunCount := 0
	action := NewRuleActionStub(&actionRunCount)
	rule.AddEnterAction(action)
	rule.AddExitAction(action)

	rule.AddCondition(&ruleCondition{`Devices[1].State`, "==", true, "uuid1234"})
	rule.AddCondition(&ruleCondition{`Devices[2].State`, "!=", "OFF", "uuid1234"})
	//rule.AddCondition(&ruleCondition{`Devices[3].State`, "!=", "OFF"})

	c := logic.ListenForChanges("uuid1234")

	node := serverprotocol.NewNode()
	state := `
		{
			"Devices": {
				"1": {
					"Id": "1",
					"Name": "Dev1",
					"State": true,
					"Type": ""
				},
				"2": {
					"Id": "2",
					"Name": "Dev2",
					"State": "ON",
					"Type": ""
				}
			}
		}
	`
	var tmp interface{}
	_ = json.Unmarshal([]byte(state), &tmp)
	node.SetState(tmp)

	logic.Update(c, node)
	//logic.SetState("uuid1234", state)
	//logic.EvaluateRules()

	state = `
		{
			"Devices": {
				"1": {
					"Id": "1",
					"Name": "Dev1",
					"State": "OFF",
					"Type": ""
				},
				"2": {
					"Id": "2",
					"Name": "Dev2",
					"State": "OFF",
					"Type": ""
				}
			}
		}
	`
	err := json.Unmarshal([]byte(state), &tmp)
	node.SetState(tmp)
	if err != nil {
		log.Println(err)
		return
	}

	logic.Update(c, node)

	// Must wait for Update to send to channel
	time.Sleep(100 * time.Millisecond)

	if len(logic.States()) != 1 {
		t.Errorf("length of logic.States should be 1. got: %s", len(logic.States()))
	}

	fmt.Println(actionRunCount)
	if actionRunCount == 2 {
		return
	}
	t.Errorf("actionRunCount wrong expected: %s got %s", 2, actionRunCount)
}

func TestParseRuleEnterExitActionsWithoutConditions(t *testing.T) {

	logic := NewLogic()

	rule := logic.AddRule("test rule without conditions")

	actionRunCount := 0
	action := NewRuleActionStub(&actionRunCount)
	rule.AddEnterAction(action)
	rule.AddExitAction(action)

	//rule.AddCondition(&ruleCondition{`Devices[1].State`, "==", false, ""})
	//rule.AddCondition(&ruleCondition{`Devices[2].State`, "!=", "OFF", ""})
	//rule.AddCondition(&ruleCondition{`Devices[3].State`, "!=", "OFF"})

	state := `
		{
			"Devices": {
				"1": {
					"Id": "1",
					"Name": "Dev1",
					"State": true,
					"Type": ""
				},
				"2": {
					"Id": "2",
					"Name": "Dev2",
					"State": "ON",
					"Type": ""
				}
			}
		}
	`

	logic.SetState("uuid1234", state)
	logic.EvaluateRules()

	state = `
		{
			"Devices": {
				"1": {
					"Id": "1",
					"Name": "Dev1",
					"State": "OFF",
					"Type": ""
				},
				"2": {
					"Id": "2",
					"Name": "Dev2",
					"State": "OFF",
					"Type": ""
				}
			}
		}
	`
	logic.SetState("uuid1234", state)
	logic.EvaluateRules()

	if len(logic.States()) != 1 {
		t.Errorf("length of logic.States should be 1. got: %s", len(logic.States()))
	}

	fmt.Println(actionRunCount)
	if actionRunCount == 0 {
		return
	}
	t.Errorf("actionRunCount wrong expected: %s got %s", 0, actionRunCount)
}
func TestEmptyRules(t *testing.T) {

	logic := NewLogic()
	logic.AddRule("test rule 1")

	assert.Len(t, logic.Rules_, 1)
	logic.EmptyRules()
	assert.Len(t, logic.Rules_, 0)
}
