package logic

import (
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/cihub/seelog"
	"github.com/jonaz/astrotime"
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

	if t.IsSunBased(t.CronWhen) != "" {
		t.reschedule()
	}
}

func (t *task) reschedule() {
	log.Debug("Rescheduling rule", t.CronWhen)

	for _, v := range t.cron.Entries() {
		if v.Id == t.CronId() {
			t.entryTime = v.Schedule.Next(time.Now().Local())
			break
		}
	}

	t.cron.RemoveFunc(t.CronId())
	t.Schedule(t.CronWhen)
}

func (t *task) Schedule(when string) {
	var err error
	t.Lock()
	t.CronWhen = when

	when = t.CalculateSun(when)

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

func (t *task) CalculateSun(when string) string {
	what := ""
	if what = t.IsSunBased(when); what == "" {
		return when
	}

	t1 := t.GetSunTime(what)
	when = strings.Replace(when, what+" "+what, strconv.Itoa(t1.Minute())+" "+strconv.Itoa(t1.Hour()), 1)
	return when
}
func (t *task) GetSunTime(what string) time.Time {

	now := time.Now()
	if !t.entryTime.IsZero() {
		now = t.entryTime
	}

	var t1 time.Time
	switch what {
	case "sunset":
		t1 = astrotime.NextSunset(now, float64(56.878333), float64(14.809167))
	case "sunrise":
		t1 = astrotime.NextSunrise(now, float64(56.878333), float64(14.809167))
	case "dusk":
		t1 = astrotime.NextDusk(now, float64(56.878333), float64(14.809167), astrotime.CIVIL_DUSK)
	case "dawn":
		t1 = astrotime.NextDawn(now, float64(56.878333), float64(14.809167), astrotime.CIVIL_DAWN)
	}
	return t1
}
