package logic

import (
	"sync"

	log "github.com/cihub/seelog"
	"github.com/elgs/cron"
	"github.com/stampzilla/stampzilla-go/stampzilla-server/protocol"
)

type task struct {
	Name     string       `json:"name"`
	Uuid_    string       `json:"uuid"`
	Actions  []RuleAction `json:"actions"`
	CronId   int
	CronWhen string
	sync.RWMutex
	nodes *protocol.Nodes
	cron  *cron.Cron
}

type Task interface {
	cron.Job
	SetUuid(string)
	Uuid() string
	AddAction(a RuleAction)
	Schedule(string)
}

func (t *task) SetUuid(uuid string) {
	t.Lock()
	t.Uuid_ = uuid
	t.Unlock()
}
func (r *task) Uuid() string {
	r.RLock()
	defer r.RUnlock()
	return r.Uuid_
}

func (t *task) Run() {
	t.RLock()
	for _, action := range t.Actions {
		action.RunCommand()
	}
	t.RUnlock()
}

func (t *task) Schedule(when string) {
	var err error
	t.Lock()
	t.CronWhen = when
	t.CronId, err = t.cron.AddJob(when, t)
	if err != nil {
		log.Error(err)
	}
	t.Unlock()
}

func (r *task) AddAction(a RuleAction) {
	if a, ok := a.(*ruleAction); ok {
		a.nodes = r.nodes
	}
	r.Lock()
	r.Actions = append(r.Actions, a)
	r.Unlock()
}
