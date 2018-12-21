package logic

import (
	"sync"

	"github.com/sirupsen/logrus"
)

//type Rule interface {
//Uuid() string
//Name() string
//SetUuid(string)
//CondState() bool
//SetCondState(bool)
//RunEnter(chan ActionProgress)
//RunExit(chan ActionProgress)
//AddExitAction(Action)
//AddEnterAction(Action)
//AddExitCancelAction(Action)
//AddEnterCancelAction(Action)
//AddCondition(RuleCondition)
//Conditions() []RuleCondition
//Operator() string
//Active() bool
//}

type Rule struct {
	Name_               string          `json:"name"`
	Uuid_               string          `json:"uuid"`
	Operator_           string          `json:"operator"`
	Active_             bool            `json:"active"`
	Enabled             bool            `json:"enabled"`
	Conditions_         []RuleCondition `json:"conditions"`
	EnterActions_       []string        `json:"enterActions"`
	ExitActions_        []string        `json:"exitActions"`
	EnterCancelActions_ []string        `json:"enterCancelActions"`
	ExitCancelActions_  []string        `json:"exitCancelActions"`
	enterActions_       []Action
	exitActions_        []Action
	enterCancelActions_ []Action
	exitCancelActions_  []Action
	condState           bool
	sync.RWMutex
}

func (r *rule) Operator() string {
	r.RLock()
	defer r.RUnlock()
	return r.Operator_
}
func (r *rule) Active() bool {
	r.RLock()
	defer r.RUnlock()
	return r.Active_
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
func (r *rule) RunEnter(progressChan chan ActionProgress) {
	logrus.Debugf("Rule enter: %s", r.Uuid())
	for _, a := range r.enterCancelActions_ {
		a.Cancel()
	}
	for _, a := range r.exitActions_ {
		a.Cancel()
	}
	for _, a := range r.enterActions_ {
		a.Run(progressChan)
	}
}
func (r *rule) RunExit(progressChan chan ActionProgress) {
	logrus.Debugf("Rule exit: %s", r.Uuid())
	for _, a := range r.exitCancelActions_ {
		a.Cancel()
	}
	for _, a := range r.enterActions_ {
		a.Cancel()
	}
	for _, a := range r.exitActions_ {
		a.Run(progressChan)
	}
}
func (r *rule) AddExitAction(a Action) {
	if a == nil {
		logrus.Error("Error adding ExitAction. Action is nil")
		return
	}
	r.Lock()
	r.exitActions_ = append(r.exitActions_, a)
	r.ExitActions_ = append(r.ExitActions_, a.Uuid())
	r.Unlock()
}
func (r *rule) AddEnterAction(a Action) {
	if a == nil {
		logrus.Error("Error adding EnterAction. Action is nil")
		return
	}
	r.Lock()
	r.enterActions_ = append(r.enterActions_, a)
	r.EnterActions_ = append(r.EnterActions_, a.Uuid())
	r.Unlock()
}
func (r *rule) AddExitCancelAction(a Action) {
	if a == nil {
		logrus.Error("Error adding ExitAction. Action is nil")
		return
	}
	r.Lock()
	r.exitCancelActions_ = append(r.exitCancelActions_, a)
	r.ExitCancelActions_ = append(r.ExitCancelActions_, a.Uuid())
	r.Unlock()
}
func (r *rule) AddEnterCancelAction(a Action) {
	if a == nil {
		logrus.Error("Error adding EnterAction. Action is nil")
		return
	}
	r.Lock()
	r.enterCancelActions_ = append(r.enterCancelActions_, a)
	r.EnterCancelActions_ = append(r.EnterCancelActions_, a.Uuid())
	r.Unlock()
}
func (r *rule) AddCondition(a RuleCondition) {
	r.Lock()
	r.Conditions_ = append(r.Conditions_, a)
	r.Unlock()
}
