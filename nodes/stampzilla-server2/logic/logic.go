package logic

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

func main() {
	fmt.Println("vim-go")
}

type Logic struct {
	Rules              []*Rule
	state              map[string]interface{}
	ActionProgressChan chan ActionProgress
	sync.RWMutex
}

func NewLogic() *Logic {
	l := &Logic{
		state:              make(map[string]interface{}),
		ActionProgressChan: make(chan ActionProgress, 100),
	}
	return l
}

func (l *Logic) AddRule(name string) *Rule {
	r := &Rule{Name_: name, Uuid_: uuid.New().String()}
	l.Lock()
	defer l.Unlock()
	l.Rules = append(l.Rules, r)
	return r
}

func (l *Logic) SetState(s map[string]interface{}) {
	l.Lock()
	l.state = s
	l.Unlock()
}

func (l *Logic) EvaluateRules() {
	for _, rule := range l.Rules {
		evaluation := l.evaluateRule(rule)
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
func (l *Logic) evaluateRule(r *Rule) bool {
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

func (l *Logic) evaluateRuleAnd(r *Rule) bool {
	for _, cond := range r.Conditions() {
		value, err := l.getValueToEvaluate(cond)
		if err != nil {
			logrus.Error(err)
			return false
		}

		if !cond.Check(value) {
			return false
		}
	}
	return true
}

func (l *Logic) evaluateRuleOr(r *Rule) bool {
	for _, cond := range r.Conditions() {
		value, err := l.getValueToEvaluate(cond)
		if err != nil {
			logrus.Error(err)
			return false
		}

		if cond.Check(value) {
			return true
		}
	}
	return false
}

func (l *Logic) getValueToEvaluate(cond RuleCondition) (interface{}, error) {
	if val, ok := l.state[cond.StatePath()]; ok {
		return val, nil
	}
	return nil, fmt.Errorf("logic: State \"%s\" not found", cond.StatePath())
}

func (l *Logic) SaveRulesToFile(path string) {
	configFile, err := os.Create(path)
	if err != nil {
		logrus.Error("creating config file", err.Error())
		return
	}
	encoder := json.NewEncoder(configFile)
	encoder.SetIndent("", "\t")
	err = encoder.Encode(l.Rules)
	if err != nil {
		logrus.Error("error marshal json", err)
	}
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
