package logic

import (
	"bytes"
	"encoding/json"
	"os"
	"sync"

	"code.google.com/p/go-uuid/uuid"
	log "github.com/cihub/seelog"
	"github.com/elgs/cron"
	"github.com/stampzilla/stampzilla-go/protocol"
	serverprotocol "github.com/stampzilla/stampzilla-go/stampzilla-server/protocol"
)

// Schedular that schedule ruleActions
type Scheduler struct {
	tasks []*task
	Nodes *serverprotocol.Nodes `inject:""`
	Cron  *cron.Cron
	sync.RWMutex
}

func NewScheduler() *Scheduler {
	scheduler := &Scheduler{}
	scheduler.Cron = cron.New()
	return scheduler
}

func (s *Scheduler) Start() {
	log.Info("Starting Scheduler")

	s.loadFromFile()
	s.Cron.Start()
}

func (s *Scheduler) AddTask(name string) Task {
	var err error

	task := &task{Name: name, Uuid_: uuid.New()}
	task.nodes = s.Nodes
	task.cron = s.Cron
	if err != nil {
		log.Error(err)
	}
	s.Lock()
	s.tasks = append(s.tasks, task)
	s.Unlock()
	return task
}

func (s *Scheduler) RemoveTask(uuid string) error {
	s.Lock()
	defer s.Unlock()
	for i, task := range s.tasks {
		if task.Uuid() == uuid {
			s.Cron.RemoveFunc(task.cronId)
			s.tasks = append(s.tasks[:i], s.tasks[i+1:]...)
			return nil
		}
	}
	return nil
}
func (s *Scheduler) CreateExampleFile() {
	task := s.AddTask("Test1")
	cmd := &protocol.Command{Cmd: "testCMD"}
	action := NewRuleAction(cmd, "simple")
	task.AddAction(action)
	task.Schedule("0 * * * * *")

	s.saveToFile()
}

func (s *Scheduler) saveToFile() {
	configFile, err := os.Create("schedule.json")
	if err != nil {
		log.Error("creating config file", err.Error())
	}
	var out bytes.Buffer
	b, err := json.Marshal(s.tasks)
	if err != nil {
		log.Error("error marshal json", err)
	}
	json.Indent(&out, b, "", "\t")
	out.WriteTo(configFile)
}

func (s *Scheduler) loadFromFile() {
	log.Info("Loading schedule from json file")

	configFile, err := os.Open("schedule.json")
	if err != nil {
		log.Error("opening config file", err.Error())
		return
	}

	type local_tasks struct {
		Name     string        `json:"name"`
		Uuid     string        `json:"uuid"`
		Actions  []*ruleAction `json:"actions"`
		CronWhen string        `json:"when"`
	}

	var tasks []*local_tasks
	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&tasks); err != nil {
		log.Error(err)
	}

	for _, task := range tasks {
		t := s.AddTask(task.Name)

		//Set the uuid from json if it exists. Otherwise use the generated one
		if task.Uuid != "" {
			t.SetUuid(task.Uuid)
		}
		for _, cond := range task.Actions {
			t.AddAction(cond)
		}
		//Schedule the task!
		t.Schedule(task.CronWhen)
	}

}
