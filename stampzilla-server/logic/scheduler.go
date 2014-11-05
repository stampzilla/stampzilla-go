package logic

import (
	"encoding/json"
	"os"
	"sync"

	"code.google.com/p/go-uuid/uuid"
	log "github.com/cihub/seelog"
	"github.com/elgs/cron"
	"github.com/stampzilla/stampzilla-go/stampzilla-server/protocol"
)

// Schedular that schedule ruleActions
type Scheduler struct {
	tasks []*task
	Nodes *protocol.Nodes `inject:""`
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
	s.tasks = append(s.tasks, task)
	return task
}

func (s *Scheduler) RemoveTask(uuid string) error {
	s.RLock()
	defer s.RUnlock()
	//Get the task
	for i, task := range s.tasks {
		if task.Uuid() == uuid {
			s.Cron.RemoveFunc(task.CronId)
			s.tasks = append(s.tasks[:i], s.tasks[i+1:]...)
			return nil
		}
	}
	//no task found!
	return nil
}

func (s *Scheduler) loadFromFile() {
	log.Info("Loading schedule from json file")

	configFile, err := os.Open("schedule.json")
	if err != nil {
		log.Error("opening config file", err.Error())
	}

	type local_tasks struct {
		Name     string        `json:"name"`
		Uuid     string        `json:"uuid"`
		Actions  []*ruleAction `json:"actions"`
		CronWhen string
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
