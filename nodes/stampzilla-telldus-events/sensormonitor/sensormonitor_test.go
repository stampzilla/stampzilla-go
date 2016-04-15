package sensormonitor

import (
	"testing"
	"time"

	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/notifications"
	"github.com/stampzilla/stampzilla-go/pkg/notifier"
	"github.com/stretchr/testify/assert"
)

type senderStub struct {
	log []string
}

func (s *senderStub) Send(msg interface{}) {
	if n, ok := msg.(notifications.Notification); ok {
		s.log = append(s.log, n.Message)
	}
}

func TestSensorDead(t *testing.T) {

	sender := &senderStub{}
	notify := notifier.New(sender)
	sm := New(notify)
	sm.MonitorSensors = []int{10}
	sm.Start()

	sm.Alive(10)
	time.Sleep(20 * time.Millisecond)

	sm.CheckDead("10ms")
	assert.Equal(t, "Sensor 10 has not been updated in 10ms", sender.log[0])
}

func TestSensorNotDead(t *testing.T) {

	sender := &senderStub{}
	notify := notifier.New(sender)
	sm := New(notify)
	sm.MonitorSensors = []int{10}
	sm.Start()

	sm.Alive(10)
	time.Sleep(10 * time.Millisecond)

	sm.CheckDead("20ms")

	assert.Empty(t, sender.log)
}
func TestNotificationSendOnlyOnce(t *testing.T) {

	sender := &senderStub{}
	notify := notifier.New(sender)
	sm := New(notify)
	sm.MonitorSensors = []int{10}
	sm.Start()

	sm.Alive(10)
	time.Sleep(20 * time.Millisecond)

	sm.CheckDead("10ms")
	sm.CheckDead("10ms")
	assert.Len(t, sender.log, 1, "Ran two CheckDead, should notify only once")

	sm.Alive(10)
	time.Sleep(20 * time.Millisecond)
	sm.CheckDead("10ms")
	assert.Len(t, sender.log, 2)
}
func TestSensorOnlyMonitorSelected(t *testing.T) {

	sender := &senderStub{}
	notify := notifier.New(sender)
	sm := New(notify)
	sm.Start()

	sm.Alive(10)
	time.Sleep(20 * time.Millisecond)

	sm.CheckDead("10ms")
	assert.Empty(t, sender.log)
}
