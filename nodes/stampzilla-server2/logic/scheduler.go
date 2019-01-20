package logic

import (
	"bytes"
	"encoding/json"
	"os"
	"sync"

	"github.com/google/uuid"
	"github.com/jonaz/cron"
	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/websocket"
)

// Scheduler that schedule running saved state actions
type Scheduler struct {
	tasks []*Task
	Cron  *cron.Cron
	sync.RWMutex

	SavedStateStore *SavedStateStore
	sender          websocket.Sender
}

func NewScheduler(savedStateStore *SavedStateStore, sender websocket.Sender) *Scheduler {
	scheduler := &Scheduler{
		SavedStateStore: savedStateStore,
		sender:          sender,
	}
	scheduler.Cron = cron.New()
	return scheduler
}

func (s *Scheduler) Start() {
	logrus.Info("Starting Scheduler")
	s.Cron.Start()
}

func (s *Scheduler) Reload() {
	//TODO verify the the new JSON is valid before unloading the existing schedule
	s.Lock()
	s.tasks = nil
	s.Unlock()
	for _, job := range s.Cron.Entries() {
		s.Cron.RemoveJob(job.Id)
	}
	s.Cron.Stop()
	s.Load()
	s.Start()
}

func (s *Scheduler) Tasks() []*Task {
	s.RLock()
	defer s.RUnlock()
	return s.tasks
}

func (s *Scheduler) AddTask(name string) *Task {
	task := &Task{
		Name_:           name,
		Uuid_:           uuid.New().String(),
		sender:          s.sender,
		SavedStateStore: s.SavedStateStore,
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
			s.Cron.RemoveJob(task.CronId())
			s.tasks = append(s.tasks[:i], s.tasks[i+1:]...)
			return nil
		}
	}
	return nil
}

func (s *Scheduler) Save() {
	configFile, err := os.Create("schedule.json")
	if err != nil {
		logrus.Error("creating config file", err.Error())
		return
	}
	var out bytes.Buffer
	b, err := json.Marshal(s.tasks)
	if err != nil {
		logrus.Error("error marshal json", err)
		return
	}
	json.Indent(&out, b, "", "\t")
	out.WriteTo(configFile)
}

func (s *Scheduler) Load() {
	logrus.Info("Loading schedule from json file")

	configFile, err := os.Open("schedule.json")
	if err != nil {
		logrus.Warn("opening config file", err.Error())
		return
	}

	s.Lock()
	defer s.Unlock()
	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&s.tasks); err != nil {
		logrus.Error(err)
		return
	}

	for _, task := range s.tasks {
		task.sender = s.sender
		task.SavedStateStore = s.SavedStateStore

		// generate uuid if missing
		if task.Uuid() == "" {
			task.SetUuid(uuid.New().String())
		}
		s.ScheduleTask(task, task.CronWhen)
	}

}

func (s *Scheduler) ScheduleTask(t *Task, when string) {
	var err error
	t.Lock()
	t.CronWhen = when

	t.cronId, err = s.Cron.AddJob(when, t)
	if err != nil {
		logrus.Error(err)
	}
	t.Unlock()
}
