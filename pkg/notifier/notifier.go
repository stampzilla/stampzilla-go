package notifier

import (
	"github.com/stampzilla/stampzilla-go/nodes/basenode"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/notifications"
)

type Notify struct {
	output basenode.Sendable
}

func New(output basenode.Sendable) *Notify {
	return &Notify{output}
}

func (n *Notify) Critical(message string) {
	n.output.Send(&notifications.Notification{
		Level:   notifications.CriticalLevel,
		Message: message,
	})
}

func (n *Notify) Error(message string) {
	n.output.Send(&notifications.Notification{
		Level:   notifications.ErrorLevel,
		Message: message,
	})
}

func (n *Notify) Warn(message string) {
	n.output.Send(&notifications.Notification{
		Level:   notifications.WarnLevel,
		Message: message,
	})
}

func (n *Notify) Info(message string) {
	n.output.Send(&notifications.Notification{
		Level:   notifications.InfoLevel,
		Message: message,
	})
}
