package logic

import "github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/notifications"

type command_notify struct {
	Notify             string                          `json:"notify"`
	Level              notifications.NotificationLevel `json:"level"`
	NotificationRouter notifications.Router            `json:"-" inject:""`
}

func (p *command_notify) Run() {
	p.NotificationRouter.Dispatch(notifications.Notification{
		Level:   p.Level,
		Message: p.Notify,
	})
}
