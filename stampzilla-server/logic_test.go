package main

import (
	"testing"

	"github.com/stampzilla/stampzilla-go/protocol"
)

func TestParseRuleEnterExitActions(t *testing.T) {

	logic := NewLogic()

	// both Devices["1"]. Devices[1] works now!!
	conditions := []*ruleCondition{
		&ruleCondition{`Devices[1].State`, "==", "OFF"},
		&ruleCondition{`Devices[2].State`, "!=", "OFF"},
	}
	enterActions := []*ruleAction{&ruleAction{&protocol.Command{"testEnterAction", nil}}}
	exitActions := []*ruleAction{&ruleAction{&protocol.Command{"testExitAction", nil}}}
	logic.AddRule("test rule 1", conditions, enterActions, exitActions)

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
