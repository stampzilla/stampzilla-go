package logic

import "testing"

func TestRuleConditionComparatorEqualBool(t *testing.T) { // {{{
	condition := &ruleCondition{`Devices[1].State`, "==", true}
	checkResult := condition.Check(true)
	if checkResult == true {
		return
	}
	t.Errorf("condition.Check(true) wrong expected: %s got %s", true, checkResult)
}                                                            // }}}
func TestRuleConditionComparatorNotEqualBool(t *testing.T) { // {{{
	condition := &ruleCondition{`Devices[1].State`, "!=", true}
	checkResult := condition.Check(true)
	if checkResult != true {
		return
	}
	t.Errorf("condition.Check(true) wrong expected: %s got %s", false, checkResult)
}                                                           // }}}
func TestRuleConditionComparatorEqualString(t *testing.T) { // {{{
	condition := &ruleCondition{`Devices[1].State`, "==", "asdf"}
	checkResult := condition.Check("asdf")
	if checkResult == true {
		return
	}
	t.Errorf("condition.Check(true) wrong expected: %s got %s", true, checkResult)
}                                                              // }}}
func TestRuleConditionComparatorNotEqualString(t *testing.T) { // {{{
	condition := &ruleCondition{`Devices[1].State`, "!=", "asdf"}
	checkResult := condition.Check("asdf")
	if checkResult != true {
		return
	}
	t.Errorf("condition.Check(true) wrong expected: %s got %s", false, checkResult)
}                                                            // }}}
func TestRuleConditionComparatorEqualBoolInt(t *testing.T) { // {{{
	condition := &ruleCondition{`Devices[1].State`, "==", 123}
	checkResult := condition.Check(123)
	if checkResult == true {
		return
	}
	t.Errorf("condition.Check(true) wrong expected: %s got %s", true, checkResult)
}                                                           // }}}
func TestRuleConditionComparatorNotEqualInt(t *testing.T) { // {{{
	condition := &ruleCondition{`Devices[1].State`, "!=", 123}
	checkResult := condition.Check(123)
	if checkResult != true {
		return
	}
	t.Errorf("condition.Check(true) wrong expected:%s got %s", false, checkResult)
}                                                                 // }}}
func TestRuleConditionComparator124GreaterThan123(t *testing.T) { // {{{
	condition := &ruleCondition{`Devices[1].State`, ">", 123}
	checkResult := condition.Check(124)
	if checkResult == true {
		return
	}
	t.Errorf("condition.Check(true) wrong expected: 124 > 123")
}                                                                    // }}}
func TestRuleConditionComparator122NotGreaterThan123(t *testing.T) { // {{{
	condition := &ruleCondition{`Devices[1].State`, ">", 123}
	checkResult := condition.Check(122)
	if checkResult != true {
		return
	}
	t.Errorf("condition.Check(true) wrong expected: 122 > 123 == false")
}                                                              // }}}
func TestRuleConditionComparator122LessThan123(t *testing.T) { // {{{
	condition := &ruleCondition{`Devices[1].State`, "<", 123}
	checkResult := condition.Check(122)
	if checkResult == true {
		return
	}
	t.Errorf("condition.Check(true) wrong expected: 122 < 123")
}                                                                 // }}}
func TestRuleConditionComparator124NotLessThan123(t *testing.T) { // {{{
	condition := &ruleCondition{`Devices[1].State`, "<", 123}
	checkResult := condition.Check(124)
	if checkResult != true {
		return
	}
	t.Errorf("condition.Check(true) wrong expected: 124 < 123 == false")
} // }}}
