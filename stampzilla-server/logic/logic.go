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

	"code.google.com/p/go-uuid/uuid"

	log "github.com/cihub/seelog"
	serverprotocol "github.com/stampzilla/stampzilla-go/stampzilla-server/protocol"
)

type Logic struct {
	states map[string]string
	Rules_ []Rule
	re     *regexp.Regexp
	sync.RWMutex
	Nodes *serverprotocol.Nodes `inject:""`
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
func (l *Logic) AddRule(name string) Rule {
	r := &rule{Name: name, Uuid_: uuid.New(), nodes: l.Nodes}
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
	configFile, err := os.Open(path)
	if err != nil {
		log.Error("opening config file", err.Error())
	}

	type local_rule struct {
		Name         string           `json:"name"`
		Uuid         string           `json:"uuid"`
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
		r := l.AddRule(rule.Name)

		//Set the uuid from json if it exists. Otherwise use the generated one
		if rule.Uuid != "" {
			r.SetUuid(rule.Uuid)
		}
		for _, cond := range rule.Conditions_ {
			r.AddCondition(cond)
		}
		for _, a := range rule.EnterActions {
			r.AddEnterAction(a)
		}
		for _, a := range rule.ExitActions {
			r.AddExitAction(a)
		}
	}
}
