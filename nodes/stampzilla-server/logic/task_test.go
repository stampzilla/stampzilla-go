package logic

import "testing"

func TestTaskSetUuid(t *testing.T) {
	task := &task{}
	task.SetUuid("test")
	if task.Uuid() == "test" {
		return
	}
	t.Errorf("Uuid wrong expected: %s got %s", "test", task.Uuid())
}

func TestTaskRunAndAddActions(t *testing.T) {
	task := &task{}
	actionRunCount := 0
	actionCancelCount := 0
	action := NewRuleActionStub(&actionRunCount, &actionCancelCount, t)

	actionService := NewActionService()
	actionService.Add(action)
	task.actionService = actionService

	task.AddAction(action)
	task.AddAction(action)
	task.Run()
	task.Run()
	if actionRunCount == 4 {
		return
	}
	t.Errorf("actionRunCount wrong expected: %d got %d", 4, actionRunCount)
}
