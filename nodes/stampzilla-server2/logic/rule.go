package logic

import (
	"encoding/json"
	"sync"
)

type Rule struct {
	Name_       string          `json:"name"`
	Uuid_       string          `json:"uuid"`
	Operator_   string          `json:"operator"`
	Active_     bool            `json:"active"`
	Enabled     bool            `json:"enabled"`
	Expression_ string          `json:"expression"`
	Conditions_ map[string]bool `json:"conditions"`
	Actions_    []string        `json:"actions"`
	//actions_    []Action
	sync.RWMutex
}

func (r *Rule) Operator() string {
	r.RLock()
	defer r.RUnlock()
	return r.Operator_
}
func (r *Rule) Expression() string {
	r.RLock()
	defer r.RUnlock()
	return r.Expression_
}
func (r *Rule) Active() bool {
	r.RLock()
	defer r.RUnlock()
	return r.Active_
}
func (r *Rule) Uuid() string {
	r.RLock()
	defer r.RUnlock()
	return r.Uuid_
}
func (r *Rule) Name() string {
	r.RLock()
	defer r.RUnlock()
	return r.Name_
}
func (r *Rule) SetUuid(uuid string) {
	r.Lock()
	r.Uuid_ = uuid
	r.Unlock()
}

func (r *Rule) Conditions() map[string]bool {
	r.RLock()
	defer r.RUnlock()
	return r.Conditions_
}
func (r *Rule) SetActive(a bool) {
	r.Lock()
	r.Active_ = a
	r.Unlock()
}

/*
func (r *Rule) RunActions(progressChan chan ActionProgress) {
	logrus.Debugf("Rule action: %s", r.Uuid())
	for _, a := range r.actions_ {
		//a.Cancel()
		a.Run(progressChan)
	}
}

// SyncActions syncronizes the action store with our actions
func (r *Rule) SyncActions(actions ActionStore) {
	r.Lock()
	r.actions_ = make([]Action, len(r.Actions_))
	for _, v := range r.Actions_ {
		r.actions_ = append(r.actions_, actions.Get(v))
	}
	r.Unlock()
}
*/

func (r *Rule) MarshalJSON() ([]byte, error) {
	r.RLock()
	defer r.RUnlock()
	type LocalRule Rule
	//TODO find a way to solve call of LocalRule copies lock value: logic.Rule (vet)
	return json.Marshal(LocalRule(*r))
}
