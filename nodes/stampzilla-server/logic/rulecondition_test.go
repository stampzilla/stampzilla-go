package logic

import "testing"

func TestRuleConditionComparatorEqualBool(t *testing.T) { // {{{
	condition := &ruleCondition{`Devices[1].State`, "==", true, "uuid1234"}
	checkResult := condition.Check(true)
	if checkResult == true {
		return
	}
	t.Errorf("condition.Check(true) wrong expected: %t got %t", true, checkResult)
} // }}}
func TestRuleConditionComparatorNotEqualBool(t *testing.T) { // {{{
	condition := &ruleCondition{`Devices[1].State`, "!=", true, "uuid1234"}
	checkResult := condition.Check(true)
	if checkResult != true {
		return
	}
	t.Errorf("condition.Check(true) wrong expected: %t got %t", false, checkResult)
} // }}}

func TestRuleConditionComparatorEqualString(t *testing.T) { // {{{
	condition := &ruleCondition{`Devices[1].State`, "==", "asdf", "uuid1234"}
	checkResult := condition.Check("asdf")
	if checkResult == true {
		return
	}
	t.Errorf("condition.Check(true) wrong expected: %t got %t", true, checkResult)
} // }}}
func TestRuleConditionComparatorNotEqualString(t *testing.T) { // {{{
	condition := &ruleCondition{`Devices[1].State`, "!=", "asdf", "uuid1234"}
	checkResult := condition.Check("asdf")
	if checkResult != true {
		return
	}
	t.Errorf("condition.Check(true) wrong expected: %t got %t", false, checkResult)
} // }}}

func TestRuleConditionComparatorEqualBoolInt(t *testing.T) { // {{{
	condition := &ruleCondition{`Devices[1].State`, "==", 123, "uuid1234"}
	checkResult := condition.Check(123)
	if checkResult == true {
		return
	}
	t.Errorf("condition.Check(true) wrong expected: %t got %t", true, checkResult)
} // }}}
func TestRuleConditionComparatorNotEqualInt(t *testing.T) { // {{{
	condition := &ruleCondition{`Devices[1].Statve`, "!=", 123, "uuid1234"}
	checkResult := condition.Check(123)
	if checkResult != true {
		return
	}
	t.Errorf("condition.Check(true) wrong expected:%t got %t", false, checkResult)
} // }}}

func TestRuleConditionComparator124GreaterThan123(t *testing.T) { // {{{
	condition := &ruleCondition{`Devices[1].State`, ">", 123, "uuid1234"}
	checkResult := condition.Check(124)
	if checkResult == true {
		return
	}
	t.Errorf("condition.Check(true) wrong expected: 124 > 123")
} // }}}
func TestRuleConditionComparator122NotGreaterThan123(t *testing.T) { // {{{
	condition := &ruleCondition{`Devices[1].State`, ">", 123, "uuid1234"}
	checkResult := condition.Check(122)
	if checkResult != true {
		return
	}
	t.Errorf("condition.Check(true) wrong expected: 122 > 123 == false")
} // }}}
func TestRuleConditionComparator122LessThan123(t *testing.T) { // {{{
	condition := &ruleCondition{`Devices[1].State`, "<", 123, "uuid1234"}
	checkResult := condition.Check(122)
	if checkResult == true {
		return
	}
	t.Errorf("condition.Check(true) wrong expected: 122 < 123")
} // }}}
func TestRuleConditionComparator124NotLessThan123(t *testing.T) { // {{{
	condition := &ruleCondition{`Devices[1].State`, "<", 123, "uuid1234"}
	checkResult := condition.Check(124)
	if checkResult != true {
		return
	}
	t.Errorf("condition.Check(true) wrong expected: 124 < 123 == false")
} // }}}

func TestRuleConditionComparator124GreaterThan123Int64(t *testing.T) { // {{{
	condition := &ruleCondition{`Devices[1].State`, ">", int64(123), "uuid1234"}
	checkResult := condition.Check(int64(124))
	if checkResult == true {
		return
	}
	t.Errorf("condition.Check(true) wrong expected: 124 > 123")
} // }}}
func TestRuleConditionComparator122NotGreaterThan123Int64(t *testing.T) { // {{{
	condition := &ruleCondition{`Devices[1].State`, ">", int64(123), "uuid1234"}
	checkResult := condition.Check(int64(122))
	if checkResult != true {
		return
	}
	t.Errorf("condition.Check(true) wrong expected: 122 > 123 == false")
} // }}}
func TestRuleConditionComparator122LessThan123Int64(t *testing.T) { // {{{
	condition := &ruleCondition{`Devices[1].State`, "<", int64(123), "uuid1234"}
	checkResult := condition.Check(int64(122))
	if checkResult == true {
		return
	}
	t.Errorf("condition.Check(true) wrong expected: 122 < 123")
} // }}}
func TestRuleConditionComparator124NotLessThan123Int64(t *testing.T) { // {{{
	condition := &ruleCondition{`Devices[1].State`, "<", int64(123), "uuid1234"}
	checkResult := condition.Check(int64(124))
	if checkResult != true {
		return
	}
	t.Errorf("condition.Check(true) wrong expected: 124 < 123 == false")
} // }}}

func TestRuleConditionComparator124GreaterThan123Float(t *testing.T) { // {{{
	condition := &ruleCondition{`Devices[1].State`, ">", 123.1, "uuid1234"}
	checkResult := condition.Check(124.1)
	if checkResult == true {
		return
	}
	t.Errorf("condition.Check(true) wrong expected: 124.1 > 123.1")
} // }}}
func TestRuleConditionComparator122NotGreaterThan123Float(t *testing.T) { // {{{
	condition := &ruleCondition{`Devices[1].State`, ">", 123.1, "uuid1234"}
	checkResult := condition.Check(122.1)
	if checkResult != true {
		return
	}
	t.Errorf("condition.Check(true) wrong expected: 122.1 > 123.1 == false")
} // }}}
func TestRuleConditionComparator122LessThan123Float(t *testing.T) { // {{{
	condition := &ruleCondition{`Devices[1].State`, "<", 123.1, "uuid1234"}
	checkResult := condition.Check(122.1)
	if checkResult == true {
		return
	}
	t.Errorf("condition.Check(true) wrong expected: 122.1 < 123")
} // }}}
func TestRuleConditionComparator124NotLessThan123Float(t *testing.T) { // {{{
	condition := &ruleCondition{`Devices[1].State`, "<", 123.1, "uuid1234"}
	checkResult := condition.Check(124.1)
	if checkResult != true {
		return
	}
	t.Errorf("condition.Check(true) wrong expected: 124.1 < 123.1 == false")
} // }}}
