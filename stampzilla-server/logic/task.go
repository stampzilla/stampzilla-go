package logic

import (
	"strings"
	"sync"
	"time"

	log "github.com/cihub/seelog"
	"github.com/jonaz/cron"
	"github.com/stampzilla/stampzilla-go/stampzilla-server/protocol"
)

type task struct {
	Name_    string       `json:"name"`
	Uuid_    string       `json:"uuid"`
	Actions  []RuleAction `json:"actions"`
	cronId   int
	CronWhen string `json:"when"`
	sync.RWMutex
	nodes     *protocol.Nodes
	cron      *cron.Cron
	entryTime time.Time
}

type Task interface {
	cron.Job
	SetUuid(string)
	Uuid() string
	Name() string
	CronId() int
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
	for _, action := range t.Actions {
		action.RunCommand()
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

func (r *task) AddAction(a RuleAction) {
	if a, ok := a.(*ruleAction); ok {
		a.nodes = r.nodes
	}
	r.Lock()
	r.Actions = append(r.Actions, a)
	r.Unlock()
}

func (t *task) IsSunBased(when string) string {
	codes := []string{
		"sunset",
		"sunrise",
		"dusk",
		"dawn",
	}

	for _, v := range codes {
		if strings.Contains(when, v) {
			return v
		}
	}
	return ""
}
