package logic

type RuleCondition interface {
	Check(interface{}) bool
	StatePath() string
	Uuid() string
}
type ruleCondition struct {
	StatePath_ string      `json:"statePath"`
	Comparator string      `json:"comparator"`
	Value      interface{} `json:"value"`
	Uuid_      string      `json:"uuid"`
}

//TODO add rlock here
func (r *ruleCondition) StatePath() string {
	return r.StatePath_
}

//TODO add rlock here
func (r *ruleCondition) Uuid() string {
	return r.Uuid_
}

func (r *ruleCondition) Check(value interface{}) bool {
	switch r.Comparator {
	case "==":
		if value == r.Value {
			return true
		}
	case "!=":
		if value != r.Value {
			return true
		}
	case "<": //less than
		return r.checkLessThan(value)
	case ">": //Greater than
		return r.checkGreaterThan(value)
	}

	return false
}

func (r *ruleCondition) checkGreaterThan(value interface{}) bool {
	switch value := value.(type) {
	case int:
		if value2, ok := r.Value.(int); ok {
			if value > value2 {
				return true
			}
		}
	case int64:
		if value2, ok := r.Value.(int64); ok {
			if value > value2 {
				return true
			}
		}
	case float64:
		if value2, ok := r.Value.(float64); ok {
			if value > value2 {
				return true
			}
		}
	}
	return false
}
func (r *ruleCondition) checkLessThan(value interface{}) bool {
	switch value := value.(type) {
	case int:
		if value2, ok := r.Value.(int); ok {
			if value < value2 {
				return true
			}
		}
	case int64:
		if value2, ok := r.Value.(int64); ok {
			if value < value2 {
				return true
			}
		}
	case float64:
		if value2, ok := r.Value.(float64); ok {
			if value < value2 {
				return true
			}
		}
	}
	return false
}
