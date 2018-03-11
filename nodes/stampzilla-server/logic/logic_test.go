package logic

import (
	"bytes"
	"encoding/json"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	log "github.com/cihub/seelog"
	serverprotocol "github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/protocol"
	"github.com/stretchr/testify/assert"
)

type ruleActionStub struct {
	actionCount *int64
	cancelCount *int64
	uuid        string
	t           *testing.T
}

func (ra *ruleActionStub) Run(c chan ActionProgress) {
	ra.t.Log("RuleActionStubRAN")
	atomic.AddInt64(ra.actionCount, 1)
	//*ra.actionCount++
}
func (ra *ruleActionStub) Cancel() {
	*ra.cancelCount++
}
func (ra *ruleActionStub) Uuid() string {
	return ra.uuid
}
func (ra *ruleActionStub) Name() string {
	return ""
}

func NewRuleActionStub(actionCount, actionCancelCount *int64, t *testing.T) *ruleActionStub {
	return &ruleActionStub{
		actionCount: actionCount,
		cancelCount: actionCancelCount,
		t:           t,
		uuid:        "",
	}
}

func TestParseRuleEnterExitActionsEvaluateTrue(t *testing.T) {

	logic := NewLogic()
	logic.Nodes = serverprotocol.NewNodes()
	node := serverprotocol.NewNode()
	node.SetName("one")
	node.SetUuid("uuid1234")
	logic.Nodes.Add(node)

	rule := logic.AddRule("test rule 1")

	actionRunCount := int64(0)
	actionCancelCount := int64(0)
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
	log.Flush()
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

	actionRunCount := int64(0)
	actionCancelCount := int64(0)
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
		t.Errorf("length of logic.States should be 1. got: %d", len(logic.States()))
	}

	if actionRunCount == 0 {
		return
	}
	t.Errorf("actionRunCount wrong expected: %d got %d", 0, actionRunCount)
	log.Flush()
}

func TestParseRuleEnterExitActionsWithoutUuid(t *testing.T) {

	logic := NewLogic()
	logic.Nodes = serverprotocol.NewNodes()
	node := serverprotocol.NewNode()
	node.SetName("one")
	node.SetUuid("uuid1234")
	logic.Nodes.Add(node)

	rule := logic.AddRule("test rule 1")

	actionRunCount := int64(0)
	actionCancelCount := int64(0)
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
	log.Flush()
}

func TestListenForChanges(t *testing.T) {

	logic := NewLogic()
	logic.Nodes = serverprotocol.NewNodes()
	node1 := serverprotocol.NewNode()
	node1.SetName("one")
	node1.SetUuid("uuid1234")
	logic.Nodes.Add(node1)

	rule := logic.AddRule("test rule 1")

	actionRunCount := int64(0)
	actionCancelCount := int64(0)
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
		t.Errorf("length of logic.States should be 1. got: %d", len(logic.States()))
	}

	if actionRunCount == 2 {
		return
	}
	t.Errorf("actionRunCount wrong expected: %d got %d", 2, actionRunCount)
	log.Flush()
}

func TestParseRuleEnterExitActionsWithoutConditions(t *testing.T) {

	logic := NewLogic()

	rule := logic.AddRule("test rule without conditions")

	actionRunCount := int64(0)
	actionCancelCount := int64(0)
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
		t.Errorf("length of logic.States should be 1. got: %d", len(logic.States()))
	}

	if actionRunCount == 0 {
		return
	}
	t.Errorf("actionRunCount wrong expected: %d got %d", 0, actionRunCount)
	log.Flush()
}
func TestEmptyRules(t *testing.T) {

	logic := NewLogic()
	logic.AddRule("test rule 1")

	assert.Len(t, logic.Rules_, 1)
	logic.EmptyRules()
	assert.Len(t, logic.Rules_, 0)
	log.Flush()
}

func TestParseRuleEnterExitActionsEvaluateTrueOperatorOr(t *testing.T) {

	logic := NewLogic()
	logic.Nodes = serverprotocol.NewNodes()
	node := serverprotocol.NewNode()
	node.SetName("one")
	node.SetUuid("uuid1234")
	logic.Nodes.Add(node)

	rule, testAction := createTestRule(t, logic, "test rule 1")
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

	assert.Equal(t, int64(1), *testAction.enter.actionCount)
	assert.Equal(t, int64(1), *testAction.exit.actionCount)
	//if *actionCancelCount != 4 {
	//t.Errorf("actionCancelCount wrong expected: %d got %d", 4, actionCancelCount)
	//}
	log.Flush()
	return
}

type test_action struct {
	enter *ruleActionStub
	exit  *ruleActionStub
}

func createTestRule(t *testing.T, l *Logic, name string) (*rule, *test_action) {
	rule := l.AddRule(name)
	actionRunCount := int64(0)
	actionCancelCount := int64(0)
	exitactionRunCount := int64(0)
	exitactionCancelCount := int64(0)
	action := &test_action{
		enter: NewRuleActionStub(&actionRunCount, &actionCancelCount, t),
		exit:  NewRuleActionStub(&exitactionRunCount, &exitactionCancelCount, t),
	}
	rule.AddEnterAction(action.enter)
	rule.AddExitAction(action.exit)
	return rule, action
}

func TestParseRuleWithRuleActiveCondition(t *testing.T) {

	var buf bytes.Buffer
	logger, err := log.LoggerFromWriterWithMinLevel(&buf, log.InfoLvl)
	assert.NoError(t, err)
	log.UseLogger(logger)

	logic := NewLogic()
	logic.Nodes = serverprotocol.NewNodes()
	node := serverprotocol.NewNode()
	node.SetName("one")
	node.SetUuid("uuid1234")
	logic.Nodes.Add(node)

	rule1, rule1Action := createTestRule(t, logic, "test rule 1")
	rule1.Operator_ = "AND"
	rule1.AddCondition(&ruleCondition{`Devices[1].State`, "==", true, "uuid1234"})
	rule1.AddCondition(&ruleCondition{`Devices[2].State`, "!=", "OFF", "uuid1234"})
	rule1.Uuid_ = "rule1uuid"

	rule2, rule2Action := createTestRule(t, logic, "test rule 2")
	rule2.Operator_ = "AND"
	rule2.AddCondition(&ruleCondition{rule1.Uuid(), "==", true, "server.logic"})
	rule2.Uuid_ = "rule2uuid"

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

	assert.Len(t, logic.States(), 1)
	assert.Equal(t, int64(1), *rule1Action.enter.actionCount)
	assert.Equal(t, int64(1), *rule1Action.exit.actionCount)
	assert.Equal(t, int64(1), *rule2Action.enter.actionCount)
	assert.Equal(t, int64(1), *rule2Action.exit.actionCount)

	log.Flush()

	//Assert the call was made in the correct order
	lines := strings.Split(buf.String(), "\n")
	assert.Contains(t, lines[0], "Rule: test rule 1 (rule1uuid) - running enter actions")
	assert.Contains(t, lines[1], "Rule: test rule 2 (rule2uuid) - running enter actions")
	assert.Contains(t, lines[2], "Rule: test rule 1 (rule1uuid) - running exit actions")
	assert.Contains(t, lines[3], "Rule: test rule 2 (rule2uuid) - running exit actions")
}
