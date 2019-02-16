package logic

import (
	"context"
	"testing"
	"time"

	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models/devices"
	"github.com/stretchr/testify/assert"
)

func TestSchedulerRunTask(t *testing.T) {

	syncer := NewMockSender()

	savedState := NewSavedStateStore()
	savedState.State["uuid"] = &SavedState{
		Name: "testname",
		UUID: "uuid",
		State: map[devices.ID]devices.State{
			devices.ID{"node", "id1"}: devices.State{
				"a": 1,
			},
		},
	}
	savedState.State["uuid2"] = &SavedState{
		Name: "testname",
		UUID: "uuid2",
		State: map[devices.ID]devices.State{
			devices.ID{"node", "id2"}: devices.State{
				"a": 2,
			},
		},
	}

	scheduler := NewScheduler(savedState, syncer)

	//Add first task.
	task := scheduler.AddTask("Test1")
	task.AddAction("uuid")
	task.SetWhen("* * * * * *")
	task.Enabled = true
	scheduler.ScheduleTask(task)

	//Add a second task.
	task = scheduler.AddTask("Test2")
	task.AddAction("uuid2")
	task.SetWhen("* * * * * *")
	task.Enabled = true
	scheduler.ScheduleTask(task)

	//Start Cron
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	scheduler.Cron.Start(ctx)
	time.Sleep(time.Second * 2)

	assert.Equal(t, int64(4), syncer.Count(), "expected to be run 4 times. 2 tasks for 2 seconds")

	assert.Equal(t, 1, syncer.Devices.Get(devices.ID{Node: "node", ID: "id1"}).State["a"])
	assert.Equal(t, 2, syncer.Devices.Get(devices.ID{Node: "node", ID: "id2"}).State["a"])
}
