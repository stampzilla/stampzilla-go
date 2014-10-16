package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"

	log "github.com/cihub/seelog"
	"github.com/stampzilla/stampzilla-go/protocol"
)

type RuleCondition interface {
	Check(string) bool
	StatePath() string
}
type ruleCondition struct {
	statePath  string
	comparator string
	value      string
}

func (r *ruleCondition) StatePath() string {
	return r.statePath
}

func (r *ruleCondition) Check(value string) bool {
	switch r.comparator {
	case "==":
		if value == r.value {
			return true
		}
	case "!=":
		if value != r.value {
			return true
		}
	}

	return false
}

type ruleAction struct {
	command *protocol.Command
	uuid    string
	nodes   *Nodes
}

func (ra *ruleAction) RunCommand() {
	fmt.Println("Running command", ra.command)
	if ra.nodes == nil {
		return
	}
	node := ra.nodes.Search(ra.uuid)
	if node != nil {
		jsonToSend, err := json.Marshal(&ra.command)
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
	stateMap map[string]string // might not be needed
	rules    []*rule
	re       *regexp.Regexp
	sync.RWMutex
}

func NewLogic() *Logic {
	l := &Logic{stateMap: make(map[string]string)}
	l.re = regexp.MustCompile(`^([^\s\[][^\s\[]*)?(\[.*?([0-9]).*?\])?$`)
	return l
}

func (l *Logic) AddRule(name string) *rule {
	r := &rule{
		name: name,
		//conditions:   conds,
		//enterActions: enteractions,
		//exitActions:  exitactions,
		condState: false,
	}
	l.Lock()
	defer l.Unlock()
	l.rules = append(l.rules, r)
	return r
}
func (l *Logic) ParseRules() {
	for _, rule := range l.rules {
		evaluation := l.parseRule(rule)
		fmt.Println("ruleEvaluationResult:", evaluation)
		if evaluation != rule.CondState() {
			rule.condState = evaluation
			if evaluation {
				rule.RunEnter()
				continue
			}
			rule.RunExit()
		}
	}
}
func (l *Logic) parseRule(r *rule) bool {
	for _, cond := range r.conditions {
		fmt.Println(cond.StatePath())
		for _, state := range l.stateMap {
			var value string
			err := l.path(state, cond.StatePath(), &value)
			if err != nil {
				log.Error(err)
			}

			// All conditions must evaluate to true
			if !cond.Check(value) {
				return false
			}
			fmt.Println("path output:", value)
		}
	}
	return true
}

func (l *Logic) ListenForChanges(uuid string) chan string {
	//TODO maybe this should be a buffered channel so we dont block on send in netStart/newClient
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
			l.ParseRules()
		}
	}
}

func (l *Logic) path(state string, jp string, t interface{}) error {
	var v interface{}
	err := json.Unmarshal([]byte(state), &v)
	if err != nil {
		return err
	}
	if jp == "" {
		return errors.New("invalid path")
	}
	for _, token := range strings.Split(jp, ".") {
		sl := l.re.FindAllStringSubmatch(token, -1)
		//fmt.Println("REGEXPtoken: ", token)
		//fmt.Println("REGEXP: ", sl)
		if len(sl) == 0 {
			return errors.New("invalid path1")
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
				return errors.New("invalid path2")
			}
			switch v2 := v.(type) {
			case []interface{}:
				v = v2[ii]
			case map[string]interface{}:
				v = v2[is]
			}
		}
	}
	rt := reflect.ValueOf(t).Elem()
	//fmt.Println("RT:", rt)
	if v == nil { //value not found
		return nil
	}
	rv := reflect.ValueOf(v)
	//fmt.Println("RT:", rv)
	rt.Set(rv)
	return nil
}

func (l *Logic) SetState(uuid, jsonData string) {
	l.Lock()
	l.stateMap[uuid] = jsonData
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
