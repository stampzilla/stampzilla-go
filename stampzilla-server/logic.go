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

	"github.com/stampzilla/stampzilla-go/protocol"
)

type ruleCondition struct {
	statePath  string
	comparator string
	value      string
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
	commands *protocol.Command
}

type rule struct {
	name         string
	conditions   []*ruleCondition
	enterActions []*ruleAction
	exitActions  []*ruleAction
	condState    bool
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

func (l *Logic) AddRule(name string, conds []*ruleCondition, enteractions []*ruleAction, exitactions []*ruleAction) {
	r := &rule{name, conds, enteractions, exitactions, false}
	l.Lock()
	l.rules = append(l.rules, r)
	l.Unlock()
}
func (l *Logic) ParseRules(states map[string]string) {
	for _, rule := range l.rules {
		evaluation := l.parseRule(states, rule)
		fmt.Println("ruleEvaluationResult:", evaluation)
		if evaluation != rule.condState {
			rule.condState = evaluation
			if evaluation {
				fmt.Println("Running Enter actions")
				continue
			}
			fmt.Println("Running Exit actions")
		}
	}
}
func (l *Logic) parseRule(s map[string]string, r *rule) bool {
	for _, cond := range r.conditions {
		fmt.Println(cond.statePath)
		for _, state := range s {
			var value string
			err := l.path(state, cond.statePath, &value)
			if err != nil {
				fmt.Println(err)
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

func (l *Logic) ListenForChanges() chan interface{} {
	//TODO maybe this should be a buffered channel so we dont block on send in netStart/newClient
	c := make(chan interface{})
	go l.listen(c)
	return c
}

// listen will run in a own goroutine and listen to incoming state changes and Parse them
func (l *Logic) listen(c chan interface{}) {
	for state := range c {
		l.ParseState(state)
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
		fmt.Println("REGEXPtoken: ", token)
		fmt.Println("REGEXP: ", sl)
		if len(sl) == 0 {
			return errors.New("invalid path1")
		}
		ss := sl[0]
		if ss[1] != "" {
			v = v.(map[string]interface{})[ss[1]]
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

func (l *Logic) ParseState(state interface{}) {
	//TODO parse all nodes.State here and generate something like this:
	// OR we dont use stateMap and only use rules Devices[2].On == true and parse it using jsonpath example below.
	// statemap["Devices[1].State"] = "OFF"
	// this might be usefull: http://play.golang.org/p/JQnry4s6KE
	// http://blog.golang.org/json-and-go
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
