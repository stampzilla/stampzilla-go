package logic

import (
	"os"
	"testing"
	"time"
)

func TestSchedulerAddTask(t *testing.T) {

	scheduler := NewScheduler()
	actionRunCount := 0
	action := NewRuleActionStub(&actionRunCount, t)

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
	action := NewRuleActionStub(&actionRunCount, t)

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

	if len(scheduler.Cron.Entries()) != 2 || len(scheduler.Tasks()) != 2 {
		t.Errorf("expected 2 cron Entries. Got: %d", len(scheduler.Cron.Entries()))
	}
	scheduler.RemoveTask(uuid1)
	if len(scheduler.Cron.Entries()) != 1 || len(scheduler.Tasks()) != 1 {
		t.Errorf("expected 1 cron Entries. Got: %d", len(scheduler.Cron.Entries()))
	}
	scheduler.RemoveTask(uuid2)
	if len(scheduler.Cron.Entries()) != 0 || len(scheduler.Tasks()) != 0 {
		t.Errorf("expected 0 cron Entries. Got: %d", len(scheduler.Cron.Entries()))
	}
}
func TestSchedulerLoadFromFile(t *testing.T) {
	scheduler := NewScheduler()
	scheduler.ActionService = NewActions()
	scheduler.loadFromFile("tests/schedule.test.json")

	if len(scheduler.Tasks()) != 1 {
		t.Errorf("expected 1 task. Got: %d", len(scheduler.Tasks()))
	}

	if scheduler.Tasks()[0].Uuid() != "7298ad6b-6827-4faa-9896-05ee61397b17" {
		t.Errorf("expected 1 task. Got: %d", len(scheduler.Tasks()))
	}

	if scheduler.Tasks()[0].Name() != "Test1" {
		t.Errorf("expected 1 task. Got: %d", len(scheduler.Tasks()))
	}

	if task, ok := scheduler.Tasks()[0].(*task); ok {
		if task.CronWhen != "0 * * * * *" {
			t.Errorf("expected 0 * * * * * . Got: %d", task.CronWhen)
		}
	} else {
		t.Error("scheudler.Tasks() schould return task compatible with *task")
	}
}
func TestSchedulersaveToFile(t *testing.T) {
	scheduler := NewScheduler()
	scheduler.ActionService = NewActions()

	actionRunCount := 0
	action := NewRuleActionStub(&actionRunCount, t)

	task1 := scheduler.AddTask("Test1")
	task1.AddAction(action)
	task1.Schedule("* * * * * *")
	uuid := task1.Uuid()
	scheduler.saveToFile("tests/schedule.json.tmp")

	scheduler = NewScheduler()
	scheduler.ActionService = NewActions()
	scheduler.loadFromFile("tests/schedule.json.tmp")

	if len(scheduler.Tasks()) != 1 {
		t.Errorf("expected 1 task. Got: %d", len(scheduler.Tasks()))
	}

	if scheduler.Tasks()[0].Uuid() != uuid {
		t.Errorf("expected 1 task. Got: %d", len(scheduler.Tasks()))
	}

	if scheduler.Tasks()[0].Name() != "Test1" {
		t.Errorf("expected 1 task. Got: %d", len(scheduler.Tasks()))
	}

	if task, ok := scheduler.Tasks()[0].(*task); ok {
		if task.CronWhen != "* * * * * *" {
			t.Errorf("expected * * * * * * . Got: %d", task.CronWhen)
		}
	} else {
		t.Error("scheudler.Tasks() schould return task compatible with *task")
	}

	os.Remove("tests/schedule.json.tmp")
}
