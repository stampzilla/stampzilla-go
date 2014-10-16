package main

import (
	"testing"

	"github.com/stampzilla/stampzilla-go/protocol"
)

func TestParseRuleEnterExitActions(t *testing.T) {

	logic := NewLogic()

	rule := logic.AddRule("test rule 1")

	rule.AddEnterAction(&ruleAction{&protocol.Command{"testEnterAction", nil}, "uuid1", nil})
	rule.AddExitAction(&ruleAction{&protocol.Command{"testExitAction", nil}, "uuid2", nil})

	rule.AddCondition(&ruleCondition{`Devices[1].State`, "==", "OFF"})
	rule.AddCondition(&ruleCondition{`Devices[2].State`, "!=", "OFF"})

	fakeStates := make(map[string]string)

	state := `
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
					"State": "ON",
					"Type": ""
				}
			}
		}
	`
	fakeStates["uuidasdfasdf"] = state

	logic.ParseRules(fakeStates)

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
	fakeStates["uuidasdfasdf"] = state

	logic.ParseRules(fakeStates)

	//t.Errorf("OutputValue wrong expected: %s got %s", 55, p.OutputValue())
}
