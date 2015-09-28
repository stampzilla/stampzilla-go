package notifier

import (
	"github.com/stampzilla/stampzilla-go/nodes/basenode"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/notifications"
)

type Notify struct {
	output basenode.Sendable
	uuid   string
	name   string
}

func New(output basenode.Sendable) *Notify {
	return &Notify{output, "", ""}
}

type Identity interface {
	Uuid() string
	Name() string
}

func (n *Notify) SetSource(src Identity) {
	n.uuid = src.Uuid()
	n.name = src.Name()
}
func (n *Notify) Critical(message string) {
	n.output.Send(notifications.Notification{
		Level:      notifications.CriticalLevel,
		Message:    message,
		Source:     n.name,
		SourceUuid: n.uuid,
	})
}

func (n *Notify) Error(message string) {
	n.output.Send(notifications.Notification{
		Level:      notifications.ErrorLevel,
		Message:    message,
		Source:     n.name,
		SourceUuid: n.uuid,
	})
}

func (n *Notify) Warn(message string) {
	n.output.Send(notifications.Notification{
		Level:      notifications.WarnLevel,
		Message:    message,
		Source:     n.name,
		SourceUuid: n.uuid,
	})
}

func (n *Notify) Info(message string) {
	n.output.Send(notifications.Notification{
		Level:      notifications.InfoLevel,
		Message:    message,
		Source:     n.name,
		SourceUuid: n.uuid,
	})
}
