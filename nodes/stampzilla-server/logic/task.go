package logic

import (
	"sync"
	"time"

	log "github.com/cihub/seelog"
	"github.com/jonaz/cron"
)

type task struct {
	Name_    string   `json:"name"`
	Uuid_    string   `json:"uuid"`
	Actions  []string `json:"actions"`
	actions  []Action
	cronId   int
	CronWhen string `json:"when"`
	sync.RWMutex
	cron      *cron.Cron
	entryTime time.Time
}

type Task interface {
	cron.Job
	SetUuid(string)
	Uuid() string
	Name() string
	CronId() int
	AddAction(a Action)
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
func (r *task) Name() string {
	r.RLock()
	defer r.RUnlock()
	return r.Name_
}
func (r *task) CronId() int {
	r.RLock()
	defer r.RUnlock()
	return r.cronId
}

func (t *task) Run() {
	t.RLock()
	for _, action := range t.actions {
		action.Run()
	}
	t.RUnlock()

}

func (t *task) Schedule(when string) {
	var err error
	t.Lock()
	t.CronWhen = when

	t.cronId, err = t.cron.AddJob(when, t)
	if err != nil {
		log.Error(err)
	}
	t.Unlock()
}

func (r *task) AddAction(a Action) {
	//if a, ok := a.(*command); ok {
	//a.nodes = r.nodes
	//}
	if a == nil {
		log.Error("Action is nil")
		return
	}
	r.Lock()
	r.actions = append(r.actions, a)
	r.Actions = append(r.Actions, a.Uuid())
	r.Unlock()
}
