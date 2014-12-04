package logic

import (
	"fmt"
	"testing"
)

type ruleActionStub struct {
	actionCount *int
}

func (ra *ruleActionStub) RunCommand() {
	fmt.Println("RuleActionStubRAN")
	*ra.actionCount++
}

func NewRuleActionStub(actionCount *int) *ruleActionStub {
	return &ruleActionStub{actionCount}
}

func TestParseRuleEnterExitActions(t *testing.T) {

	logic := NewLogic()

	rule := logic.AddRule("test rule 1")

	//rule.AddEnterAction(&ruleAction{&protocol.Command{"testEnterAction", nil}, "uuid1", nil})
	//rule.AddExitAction(&ruleAction{&protocol.Command{"testExitAction", nil}, "uuid2", nil})

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

	fmt.Println(actionRunCount)
	if actionRunCount == 2 {
		return
	}
	t.Errorf("actionRunCount wrong expected: %s got %s", 2, actionRunCount)
}
