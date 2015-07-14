package logic

import (
	"sync"

	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/protocol"
)

type Rule interface {
	Uuid() string
	Name() string
	SetUuid(string)
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
	Name_         string          `json:"name"`
	Uuid_         string          `json:"uuid"`
	Conditions_   []RuleCondition `json:"conditions"`
	EnterActions_ []RuleAction    `json:"enterActions"`
	ExitActions_  []RuleAction    `json:"exitActions"`
	condState     bool
	sync.RWMutex
	nodes *protocol.Nodes
}

func (r *rule) Uuid() string {
	r.RLock()
	defer r.RUnlock()
	return r.Uuid_
}
func (r *rule) Name() string {
	r.RLock()
	defer r.RUnlock()
	return r.Name_
}
func (r *rule) SetUuid(uuid string) {
	r.Lock()
	r.Uuid_ = uuid
	r.Unlock()
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
	if a, ok := a.(*ruleAction); ok {
		a.nodes = r.nodes
	}
	r.Lock()
	r.ExitActions_ = append(r.ExitActions_, a)
	r.Unlock()
}
func (r *rule) AddEnterAction(a RuleAction) {
	if a, ok := a.(*ruleAction); ok {
		a.nodes = r.nodes
	}
	r.Lock()
	r.EnterActions_ = append(r.EnterActions_, a)
	r.Unlock()
}
func (r *rule) AddCondition(a RuleCondition) {
	r.Lock()
	r.Conditions_ = append(r.Conditions_, a)
	r.Unlock()
}
