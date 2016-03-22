package logic

import (
	"testing"

	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/notifications"
)

type fakeRouter struct {
	notifications []notifications.Notification
}

func (fr *fakeRouter) Dispatch(msg notifications.Notification) {
	fr.notifications = append(fr.notifications, msg)
}

func (fr *fakeRouter) Send(data interface{}) {
}

func TestCommandNotifyRun(t *testing.T) {

	fakeRouter := &fakeRouter{}

	c := &command_notify{
		Notify: "Testing",
		Level:  notifications.ErrorLevel,
	}
	c.NotificationRouter = fakeRouter

	c.Run()

	if fakeRouter.notifications[0].Message != "Testing" {
		t.Errorf("Got %s Expected %s", fakeRouter.notifications[0].Message, "Testing")
	}
	if fakeRouter.notifications[0].Level != notifications.ErrorLevel {
		t.Errorf("Got %s Expected %s", fakeRouter.notifications[0].Level, notifications.ErrorLevel)
	}
}
