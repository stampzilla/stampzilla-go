package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"

	log "github.com/cihub/seelog"
	"github.com/stampzilla/stampzilla-go/protocol"
)

type RuleCondition interface {
	Check(interface{}) bool
	StatePath() string
}
type ruleCondition struct {
	StatePath_ string      `json:"statePath"`
	Comparator string      `json:"comparator"`
	Value      interface{} `json:"value"`
}

func (r *ruleCondition) StatePath() string {
	return r.StatePath_
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
	case "<":
		//TODO here we need to do type assertsion so that we can only compare int and float i think!
		//if value < r.Value {
		//return true
		//}
	case ">":
		//TODO here we need to do type assertsion so that we can only compare int and float i think!
		//if value < r.Value {
		//return true
		//}
	}

	return false
}

type ruleAction struct {
	Command *protocol.Command
	Uuid    string
	Nodes   *Nodes
}

func (ra *ruleAction) RunCommand() {
	fmt.Println("Running command", ra.Command)
	if ra.Nodes == nil {
		return
	}
	node := ra.Nodes.Search(ra.Uuid)
	if node != nil {
		jsonToSend, err := json.Marshal(&ra.Command)
		if err != nil {
			log.Error(err)
			return
		}
		node.conn.Write(jsonToSend)
	}
}

type RuleAction interface {
	RunCommand()
}

type Rule interface {
	CondState() bool
	SetCondState(bool)
	RunEnter()
	RunExit()
	AddExitAction(RuleAction)
	AddEnterAction(RuleAction)
	AddCondition(RuleCondition)
	Conditions() []RuleCondition
}

type rule struct {
	name         string
	conditions   []RuleCondition
	enterActions []RuleAction
	exitActions  []RuleAction
	condState    bool
	sync.RWMutex
}

func (r *rule) CondState() bool {
	r.RLock()
	defer r.RUnlock()
	return r.condState
}
func (r *rule) Conditions() []RuleCondition {
	r.RLock()
	defer r.RUnlock()
	return r.conditions
}

func (r *rule) SetCondState(cond bool) {
	r.RLock()
	r.condState = cond
	r.RUnlock()
}
func (r *rule) RunEnter() {
	for _, a := range r.enterActions {
		a.RunCommand()
	}
}
func (r *rule) RunExit() {
	for _, a := range r.exitActions {
		a.RunCommand()
	}
}
func (r *rule) AddExitAction(a RuleAction) {
	r.Lock()
	r.exitActions = append(r.exitActions, a)
	r.Unlock()
}
func (r *rule) AddEnterAction(a RuleAction) {
	r.Lock()
	r.enterActions = append(r.enterActions, a)
	r.Unlock()
}
func (r *rule) AddCondition(a RuleCondition) {
	r.Lock()
	r.conditions = append(r.conditions, a)
	r.Unlock()
}

type Logic struct {
	states map[string]string
	rules  []Rule
	re     *regexp.Regexp
	sync.RWMutex
}

func NewLogic() *Logic {
	l := &Logic{states: make(map[string]string)}
	l.re = regexp.MustCompile(`^([^\s\[][^\s\[]*)?(\[.*?([0-9]).*?\])?$`)
	return l
}

func (l *Logic) States() map[string]string {
	l.RLock()
	defer l.RUnlock()
	return l.states
}
func (l *Logic) AddRule(name string) Rule {
	r := &rule{name: name}
	l.Lock()
	defer l.Unlock()
	l.rules = append(l.rules, r)
	return r
}
func (l *Logic) EvaluateRules() {
	for _, rule := range l.rules {
		evaluation := l.evaluateRule(rule)
		fmt.Println("ruleEvaluationResult:", evaluation)
		if evaluation != rule.CondState() {
			rule.SetCondState(evaluation)
			if evaluation {
				rule.RunEnter()
				continue
			}
			rule.RunExit()
		}
	}
}
func (l *Logic) evaluateRule(r Rule) bool {
	for _, cond := range r.Conditions() {
		fmt.Println(cond.StatePath())
		for _, state := range l.States() {
			//var value string
			value, err := l.path(state, cond.StatePath())
			if err != nil {
				log.Error(err)
			}

			fmt.Println("path output:", value)
			// All conditions must evaluate to true
			if !cond.Check(value) {
				return false
			}
		}
	}
	return true
}

func (l *Logic) ListenForChanges(uuid string) chan string {
	c := make(chan string)
	go l.listen(uuid, c)
	return c
}

// listen will run in a own goroutine and listen to incoming state changes and Parse them
func (l *Logic) listen(uuid string, c chan string) {
	for {
		select {
		case state, open := <-c:
			if !open {
				return
			}
			l.SetState(uuid, state)
			l.EvaluateRules()
		}
	}
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
	//rt := reflect.ValueOf(t).Elem()
	////fmt.Println("RT:", rt)
	//if v == nil { //value not found
	//return nil, nil
	//}
	//rv := reflect.ValueOf(v)
	//switch vv := v.(type) {
	//case bool:
	////this doesnt work yet!
	//if vv {
	//t = "true"
	//return nil, nil
	//}
	////rt.Set("false")
	//t = "false"
	//fmt.Println("ITS A BOOL")
	//case string:
	//rt.Set(rv)
	//}
}

func (l *Logic) SetState(uuid, jsonData string) {
	l.Lock()
	l.states[uuid] = jsonData
	l.Unlock()
}

/*
Example of state:
State: {
	Devices: {
		1: {
			Id: "1",
			Name: "Dev1",
			State: "OFF",
			Type: ""
		},
		2: {
			Id: "2",
			Name: "Dev2",
			State: "ON",
			Type: ""
		}
	}
}
*/
