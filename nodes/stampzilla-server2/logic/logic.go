package logic

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

func main() {
	fmt.Println("vim-go")
}

type Logic struct {
	Rules []*Rule
}

func NewLogic() *Logic {
	l := &Logic{states: make(map[string]string)}
	l.re = regexp.MustCompile(`^([^\s\[][^\s\[]*)?(\[.*?([0-9]+).*?\])?$`)
	return l
}

func (l *Logic) EvaluateRules() {
	for _, rule := range l.Rules() {
		evaluation := l.evaluateRule(rule)
		//fmt.Println("ruleEvaluationResult:", evaluation)
		if evaluation != rule.CondState() {
			rule.SetCondState(evaluation)
			if evaluation {
				logrus.Info("Rule: ", rule.Name(), " (", rule.Uuid(), ") - running enter actions")
				rule.RunEnter(l.ActionProgressChan)
				rule.Lock()
				rule.Active_ = true
				rule.Unlock()
				continue
			}

			logrus.Info("Rule: ", rule.Name(), " (", rule.Uuid(), ") - running exit actions")
			rule.RunExit(l.ActionProgressChan)
			rule.Lock()
			rule.Active_ = false
			rule.Unlock()
		}
	}
}
func (l *Logic) evaluateRule(r Rule) bool {
	//TODO if rule.enabled is false return false here??? default enabled to true in existing rules or do migration?
	if len(r.Conditions()) == 0 {
		return false
	}

	if strings.ToLower(r.Operator()) == "and" || r.Operator() == "" {
		return l.evaluateRuleAnd(r)
	}
	if strings.ToLower(r.Operator()) == "or" {
		return l.evaluateRuleOr(r)
	}

	return false
}

func (l *Logic) path(state string, jp string) (interface{}, error) {
	var v interface{}
	err := json.Unmarshal([]byte(state), &v)
	if err != nil {
		return nil, err
	}
	if jp == "" {
		return nil, errors.New("invalid path")
	}
	for _, token := range strings.Split(jp, ".") {
		sl := l.re.FindAllStringSubmatch(token, -1)
		//fmt.Println("REGEXPtoken: ", token)
		//fmt.Println("REGEXP: ", sl)
		if len(sl) == 0 {
			return nil, errors.New("invalid path1")
		}
		ss := sl[0]
		if ss[1] != "" {
			switch v1 := v.(type) {
			case map[string]interface{}:
				v = v1[ss[1]]
			}
		}
		if ss[3] != "" {
			ii, err := strconv.Atoi(ss[3])
			is := ss[3]
			if err != nil {
				return nil, errors.New("invalid path2")
			}
			switch v2 := v.(type) {
			case []interface{}:
				v = v2[ii]
			case map[string]interface{}:
				v = v2[is]
			}
		}
	}
	return v, nil
}

/*
old way:

{
	"name": "All off",
	"enabled": true,
	"uuid": "e8092b86-1261-44cd-ab64-38121df58a79",
	"conditions": [
		{
			"statePath": "Devices.fefe749b.Status",
			"comparator": "==",
			"value": "R1B0",
			"uuid": "efd2bd24-ac50-4147-bdf9-da3dd12c8f8a"
		}
	],
	"enterActions": [
		"c7d352bb-23f4-468c-b476-f76599c09a0d"
	]
},
*/
