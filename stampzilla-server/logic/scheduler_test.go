package logic

import (
	"testing"
	"time"
)

func TestSchedulerAddTask(t *testing.T) {

	scheduler := NewScheduler()
	actionRunCount := 0
	action := NewRuleActionStub(&actionRunCount)

	//Add first task.
	task := scheduler.AddTask("Test1")
	task.AddAction(action)
	task.Schedule("* * * * * *")

	//Add a second task.
	task = scheduler.AddTask("Test2")
	task.AddAction(action)
	task.Schedule("* * * * * *")

	//Start Cron
	scheduler.Cron.Start()
	//Wait!
	time.Sleep(time.Second * 2)

	if actionRunCount == 4 {
		return
	}
	t.Errorf("actionRunCount wrong expected cron to have ran 4 times got %d", actionRunCount)
}
