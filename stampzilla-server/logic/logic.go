package logic

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"

	log "github.com/cihub/seelog"
	"github.com/stampzilla/stampzilla-go/protocol"
	serverprotocol "github.com/stampzilla/stampzilla-go/stampzilla-server/protocol"
)

var nodes *serverprotocol.Nodes

type RuleCondition interface {
	Check(interface{}) bool
	StatePath() string
}
type ruleCondition struct {
	StatePath_ string      `json:"statePath"`
	Comparator string      `json:"comparator"`
	Value      interface{} `json:"value"`
}

func NewRuleCondition(path, comp string, val interface{}) RuleCondition {
	return &ruleCondition{path, comp, val}
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
	Command *protocol.Command `json:"command"`
	Uuid    string            `json:"uuid"`
}

func NewRuleAction(cmd *protocol.Command, uuid string) RuleAction {
	return &ruleAction{cmd, uuid}
}

func (ra *ruleAction) RunCommand() {
	fmt.Println("Running command", ra.Command)
	if nodes == nil {
		return
	}
	node := nodes.Search(ra.Uuid)
	if node != nil {
		jsonToSend, err := json.Marshal(&ra.Command)
		if err != nil {
			log.Error(err)
			return
		}
		node.Conn().Write(jsonToSend)
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
	Name          string          `json:"name"`
	Conditions_   []RuleCondition `json:"conditions"`
	EnterActions_ []RuleAction    `json:"enterActions"`
	ExitActions_  []RuleAction    `json:"exitActions"`
	condState     bool
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
	return r.Conditions_
}

func (r *rule) SetCondState(cond bool) {
	r.RLock()
	r.condState = cond
	r.RUnlock()
}
func (r *rule) RunEnter() {
	for _, a := range r.EnterActions_ {
		a.RunCommand()
	}
}
func (r *rule) RunExit() {
	for _, a := range r.ExitActions_ {
		a.RunCommand()
	}
}
func (r *rule) AddExitAction(a RuleAction) {
	r.Lock()
	r.ExitActions_ = append(r.ExitActions_, a)
	r.Unlock()
}
func (r *rule) AddEnterAction(a RuleAction) {
	r.Lock()
	r.EnterActions_ = append(r.EnterActions_, a)
	r.Unlock()
}
func (r *rule) AddCondition(a RuleCondition) {
	r.Lock()
	r.Conditions_ = append(r.Conditions_, a)
	r.Unlock()
}

type Logic struct {
	states map[string]string
	Rules_ []Rule
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
	r := &rule{Name: name}
	l.Lock()
	defer l.Unlock()
	l.Rules_ = append(l.Rules_, r)
	return r
}
func (l *Logic) EvaluateRules() {
	for _, rule := range l.Rules() {
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
}

func (l *Logic) SetNodes(n *serverprotocol.Nodes) {
	nodes = n
}
func (l *Logic) SetState(uuid, jsonData string) {
	l.Lock()
	l.states[uuid] = jsonData
	l.Unlock()
}
func (l *Logic) Rules() []Rule {
	l.RLock()
	defer l.RUnlock()
	return l.Rules_
}

func (l *Logic) SaveRulesToFile(path string) {
	configFile, err := os.Create(path)
	if err != nil {
		log.Error("creating config file", err.Error())
	}
	var out bytes.Buffer
	b, err := json.MarshalIndent(l.Rules, "", "\t")
	if err != nil {
		log.Error("error marshal json", err)
	}
	json.Indent(&out, b, "", "\t")
	out.WriteTo(configFile)
}

func (l *Logic) RestoreRulesFromFile(path string) {
	//TODO finish this. We have to implement UnmarshalJSON([]byte) error on all our interfaces
	// in order for json Deocode to work.
	configFile, err := os.Open(path)
	if err != nil {
		log.Error("opening config file", err.Error())
	}

	type local_rule struct {
		Name         string           `json:"name"`
		Conditions_  []*ruleCondition `json:"conditions"`
		EnterActions []*ruleAction    `json:"enterActions"`
		ExitActions  []*ruleAction    `json:"exitActions"`
	}

	var rules []*local_rule
	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&rules); err != nil {
		log.Error(err)
	}

	for _, rule := range rules {
		newRule := l.AddRule(rule.Name)
		for _, newCond := range rule.Conditions_ {
			newRule.AddCondition(newCond)
		}
		for _, newEnterAction := range rule.EnterActions {
			newRule.AddEnterAction(newEnterAction)
		}
		for _, newExtiAction := range rule.ExitActions {
			newRule.AddExitAction(newExtiAction)
		}
	}

}
