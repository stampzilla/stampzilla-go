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
func TestSchedulerRemoveTask(t *testing.T) {

	scheduler := NewScheduler()
	actionRunCount := 0
	action := NewRuleActionStub(&actionRunCount)

	//Add first task.
	task := scheduler.AddTask("Test1")
	task.AddAction(action)
	task.Schedule("* * * * * *")
	uuid1 := task.Uuid()

	//Add a second task.
	task = scheduler.AddTask("Test2")
	task.AddAction(action)
	task.Schedule("* * * * * *")
	uuid2 := task.Uuid()

	//Start Cron
	scheduler.Cron.Start()

	if len(scheduler.Cron.Entries()) != 2 {
		t.Errorf("expected 2 cron Entries. Got: %d", len(scheduler.Cron.Entries()))
	}
	scheduler.RemoveTask(uuid1)
	if len(scheduler.Cron.Entries()) != 1 {
		t.Errorf("expected 1 cron Entries. Got: %d", len(scheduler.Cron.Entries()))
	}
	scheduler.RemoveTask(uuid2)
	if len(scheduler.Cron.Entries()) != 0 {
		t.Errorf("expected 0 cron Entries. Got: %d", len(scheduler.Cron.Entries()))
	}
}
