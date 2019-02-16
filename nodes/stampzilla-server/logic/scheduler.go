package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/google/uuid"
	"github.com/jonaz/cron"
	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/websocket"
)

/* schedule.json
{
    "398d8a42-cb31-42ff-b4ae-99bdbb44d0ae": {
        "name": "test1",
        "uuid": "398d8a42-cb31-42ff-b4ae-99bdbb44d0ae",
        "when": "0 * * * * *",
        "actions": [
            "6fbaea24-6b3f-4856-9194-735b349bbf4d"
        ]
    }
}
*/

type Tasks map[string]*Task

// Scheduler that schedule running saved state actions
type Scheduler struct {
	tasks Tasks
	Cron  *cron.Cron
	sync.RWMutex

	SavedStateStore *SavedStateStore
	sender          websocket.Sender
	stop            context.CancelFunc
}

func NewScheduler(savedStateStore *SavedStateStore, sender websocket.Sender) *Scheduler {
	scheduler := &Scheduler{
		SavedStateStore: savedStateStore,
		sender:          sender,
		tasks:           make(Tasks),
	}
	scheduler.Cron = cron.New()
	return scheduler
}

func (s *Scheduler) Start(parentCtx context.Context) {
	ctx, cancel := context.WithCancel(parentCtx)
	s.Lock()
	s.stop = cancel
	s.Cron.Start(ctx)
	s.Unlock()
	logrus.Info("scheduler: started")
}

func (s *Scheduler) Stop() {
	s.Lock()
	if s.stop != nil {
		s.stop()
	}
	s.Unlock()
	logrus.Info("Scheduler stopped")
}

func (s *Scheduler) SetTasks(t Tasks) {
	s.Lock()
	s.tasks = t
	s.syncTaskDependencies()
	s.Unlock()
}
func (s *Scheduler) Tasks() Tasks {
	s.RLock()
	defer s.RUnlock()
	return s.tasks
}

func (s *Scheduler) AddTask(name string) *Task {
	task := &Task{
		XName:           name,
		XUuid:           uuid.New().String(),
		sender:          s.sender,
		savedStateStore: s.SavedStateStore,
	}
	s.Lock()
	s.tasks[task.XUuid] = task
	s.Unlock()
	return task
}
func (s *Scheduler) Task(uuid string) *Task {
	s.RLock()
	defer s.RUnlock()
	return s.tasks[uuid]
}

func (s *Scheduler) RemoveTask(uuid string) {
	s.Lock()
	defer s.Unlock()
	task := s.Task(uuid)
	if task == nil {
		return
	}
	s.Cron.RemoveJob(task.CronId())
	delete(s.tasks, uuid)
}

func (s *Scheduler) Save() error {
	configFile, err := os.Create("schedule.json")
	if err != nil {
		return fmt.Errorf("scheduler: error saving tasks: %s", err)
	}
	encoder := json.NewEncoder(configFile)
	encoder.SetIndent("", "\t")
	s.Lock()
	defer s.Unlock()
	err = encoder.Encode(s.tasks)
	if err != nil {
		return fmt.Errorf("scheduler: error saving tasks: %s", err)
	}
	return nil
}

func (s *Scheduler) Load() error {
	logrus.Info("Loading schedule from json file")

	configFile, err := os.Open("schedule.json")
	if err != nil {
		if os.IsNotExist(err) {
			logrus.Warn(err)
			return nil // We dont want to error our if the file does not exist when we start the server
		}
		return err
	}

	s.Lock()
	defer s.Unlock()
	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&s.tasks); err != nil {
		return fmt.Errorf("scheduler: error loading tasks: %s", err)
	}

	s.syncTaskDependencies()

	return nil
}

func (s *Scheduler) syncTaskDependencies() {
	s.Cron.RemoveAll() // always remove all when we load new tasks to make sure we dont get duplicate jobs scheduled
	for _, task := range s.tasks {
		task.Lock()
		if task.sender == nil {
			task.sender = s.sender
		}
		if task.savedStateStore == nil {
			task.savedStateStore = s.SavedStateStore
		}
		task.Unlock()

		// generate uuid if missing
		if task.Uuid() == "" {
			task.SetUuid(uuid.New().String())
		}
		s.ScheduleTask(task)
	}

}

func (s *Scheduler) ScheduleTask(t *Task) {
	var err error
	t.Lock()
	t.cronID, err = s.Cron.AddJob(t.When, t)
	if err != nil {
		logrus.Error(err)
	}
	t.Unlock()
}
