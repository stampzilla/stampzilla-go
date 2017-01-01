package logic

import (
	"sync"
	"time"

	log "github.com/cihub/seelog"
	"github.com/jonaz/cron"
	"github.com/stampzilla/stampzilla-go/protocol"
)

type task struct {
	Name_   string   `json:"name"`
	Uuid_   string   `json:"uuid"`
	Actions []string `json:"actions"`
	//actions  []Action
	cronId   int
	CronWhen string `json:"when"`
	sync.RWMutex
	cron      *cron.Cron
	entryTime time.Time

	ActionProgressChan chan ActionProgress `json:"-"`
	actionService      *ActionService      `json:"-"`
}

type Task interface {
	cron.Job
	SetUuid(string)
	Uuid() string
	Name() string
	CronId() int
	AddAction(a protocol.Identifiable)
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
	for _, uuid := range t.Actions {
		a := t.actionService.GetByUuid(uuid)
		if a == nil {
			log.Errorf("Action %s not found in actionService", uuid)
			continue
		}
		a.Run(t.ActionProgressChan)
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

func (r *task) AddAction(i protocol.Identifiable) {
	r.Lock()
	r.Actions = append(r.Actions, i.Uuid())
	r.Unlock()
}
