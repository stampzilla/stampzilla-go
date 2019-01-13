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

	s.loadFromFile("schedule.json")
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

/*
func (s *Scheduler) CreateExampleFile() {
	task := s.AddTask("Test1")
	cmd := &protocol.Command{Cmd: "testCMD"}
	action := &action{
		Commands: []Command{NewCommand(cmd, "simple")},
	}
	task.AddAction(action)
	task.Schedule("0 * * * * *")

	s.saveToFile("schedule.json")
}
*/

func (s *Scheduler) saveToFile(filepath string) {
	configFile, err := os.Create(filepath)
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

func (s *Scheduler) loadFromFile(filepath string) {
	logrus.Info("Loading schedule from json file")

	configFile, err := os.Open(filepath)
	if err != nil {
		logrus.Warn("opening config file", err.Error())
		return
	}

	//TODO we can use normal task here if we refactor some stuff :)
	type localTasks struct {
		Name     string   `json:"name"`
		UUID     string   `json:"uuid"`
		Actions  []string `json:"actions"`
		CronWhen string   `json:"when"`
	}

	var tasks []*localTasks
	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&tasks); err != nil {
		logrus.Error(err)
		return
	}

	for _, task := range tasks {
		t := s.AddTask(task.Name)

		//Set the uuid from json if it exists. Otherwise use the generated one
		if task.UUID != "" {
			t.SetUuid(task.UUID)
		}
		for _, uuid := range task.Actions {
			//a := s.ActionService.GetByUuid(uuid)
			//if a == nil {
			//logrus.Errorf("Could not find action %s. Skipping adding it to task.\n", uuid)
			//continue
			//}
			t.AddAction(uuid)
		}
		//Schedule the task!
		s.ScheduleTask(t, task.CronWhen)
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
