package logic

import (
	"encoding/json"
	"testing"
	"time"

	serverprotocol "github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/protocol"
	"github.com/stretchr/testify/assert"
)

type ruleActionStub struct {
	actionCount *int
	cancelCount *int
	t           *testing.T
}

func (ra *ruleActionStub) Run() {
	ra.t.Log("RuleActionStubRAN")
	*ra.actionCount++
}
func (ra *ruleActionStub) Cancel() {
	*ra.cancelCount++
}
func (ra *ruleActionStub) Uuid() string {
	return ""
}
func (ra *ruleActionStub) Name() string {
	return ""
}

func NewRuleActionStub(actionCount, actionCancelCount *int, t *testing.T) *ruleActionStub {
	return &ruleActionStub{actionCount, actionCancelCount, t}
}

func TestParseRuleEnterExitActionsEvaluateTrue(t *testing.T) {

	logic := NewLogic()
	logic.Nodes = serverprotocol.NewNodes()
	node := serverprotocol.NewNode()
	node.SetName("one")
	node.SetUuid("uuid1234")
	logic.Nodes.Add(node)

	rule := logic.AddRule("test rule 1")

	actionRunCount := 0
	actionCancelCount := 0
	action := NewRuleActionStub(&actionRunCount, &actionCancelCount, t)
	rule.AddEnterAction(action)
	rule.AddExitAction(action)
	rule.AddEnterCancelAction(action)
	rule.AddExitCancelAction(action)

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

	assert.Equal(t, true, rule.Active())

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

	assert.Equal(t, false, rule.Active())

	if len(logic.States()) != 1 {
		t.Errorf("length of logic.States should be 1. got: %d", len(logic.States()))
	}

	if actionRunCount != 2 {
		t.Errorf("actionRunCount wrong expected: %d got %d", 2, actionRunCount)
	}
	if actionCancelCount != 4 {
		t.Errorf("actionCancelCount wrong expected: %d got %d", 4, actionCancelCount)
	}
	return
}

func TestParseRuleEnterExitActionsEvaluateFalse(t *testing.T) {

	logic := NewLogic()
	logic.Nodes = serverprotocol.NewNodes()
	node := serverprotocol.NewNode()
	node.SetName("one")
	node.SetUuid("uuid1234")
	logic.Nodes.Add(node)

	rule := logic.AddRule("test rule 1")

	actionRunCount := 0
	actionCancelCount := 0
	action := NewRuleActionStub(&actionRunCount, &actionCancelCount, t)
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

	if actionRunCount == 0 {
		return
	}
	t.Errorf("actionRunCount wrong expected: %s got %s", 0, actionRunCount)
}

func TestParseRuleEnterExitActionsWithoutUuid(t *testing.T) {

	logic := NewLogic()
	logic.Nodes = serverprotocol.NewNodes()
	node := serverprotocol.NewNode()
	node.SetName("one")
	node.SetUuid("uuid1234")
	logic.Nodes.Add(node)

	rule := logic.AddRule("test rule 1")

	actionRunCount := 0
	actionCancelCount := 0
	action := NewRuleActionStub(&actionRunCount, &actionCancelCount, t)
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
		t.Errorf("length of logic.States should be 1. got: %d", len(logic.States()))
	}

	if actionRunCount == 0 {
		return
	}
	t.Errorf("actionRunCount wrong expected: %d got %d", 0, actionRunCount)
}

func TestListenForChanges(t *testing.T) {

	logic := NewLogic()
	logic.Nodes = serverprotocol.NewNodes()
	node1 := serverprotocol.NewNode()
	node1.SetName("one")
	node1.SetUuid("uuid1234")
	logic.Nodes.Add(node1)

	rule := logic.AddRule("test rule 1")

	actionRunCount := 0
	actionCancelCount := 0
	action := NewRuleActionStub(&actionRunCount, &actionCancelCount, t)
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
		t.Error(err)
		return
	}

	logic.Update(c, node)

	// Must wait for Update to send to channel
	time.Sleep(100 * time.Millisecond)

	if len(logic.States()) != 1 {
		t.Errorf("length of logic.States should be 1. got: %s", len(logic.States()))
	}

	if actionRunCount == 2 {
		return
	}
	t.Errorf("actionRunCount wrong expected: %s got %s", 2, actionRunCount)
}

func TestParseRuleEnterExitActionsWithoutConditions(t *testing.T) {

	logic := NewLogic()

	rule := logic.AddRule("test rule without conditions")

	actionRunCount := 0
	actionCancelCount := 0
	action := NewRuleActionStub(&actionRunCount, &actionCancelCount, t)
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

func TestParseRuleEnterExitActionsEvaluateTrueOperatorOr(t *testing.T) {

	logic := NewLogic()
	logic.Nodes = serverprotocol.NewNodes()
	node := serverprotocol.NewNode()
	node.SetName("one")
	node.SetUuid("uuid1234")
	logic.Nodes.Add(node)

	rule := logic.AddRule("test rule 1")

	actionRunCount := 0
	actionCancelCount := 0
	action := NewRuleActionStub(&actionRunCount, &actionCancelCount, t)
	rule.AddEnterAction(action)
	rule.AddExitAction(action)
	rule.AddEnterCancelAction(action)
	rule.AddExitCancelAction(action)
	rule.Operator_ = "OR"

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
					"State": "OFF",
					"Type": ""
				}
			}
		}
	`

	logic.SetState("uuid1234", state)
	logic.EvaluateRules()

	assert.Equal(t, true, rule.Active())

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

	assert.Equal(t, false, rule.Active())

	if len(logic.States()) != 1 {
		t.Errorf("length of logic.States should be 1. got: %d", len(logic.States()))
	}

	if actionRunCount != 2 {
		t.Errorf("actionRunCount wrong expected: %d got %d", 2, actionRunCount)
	}
	if actionCancelCount != 4 {
		t.Errorf("actionCancelCount wrong expected: %d got %d", 4, actionCancelCount)
	}
	return
}
