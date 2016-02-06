package logic

import (
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/notifications"
	serverprotocol "github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/protocol"
)

type command_notify struct {
	Notify             string                          `json:"notify"`
	Level              notifications.NotificationLevel `json:"level"`
	NotificationRouter *notifications.Router           `json:"-" inject:""`
}

func NewNotify(duration string) *command_notify {
	p := &command_notify{}
	return p
}

func (p *command_notify) Run() {
	//notify := notifier.New(p.NotificationRouter)
	//notify.Critical
	p.NotificationRouter.Send(notifications.Notification{
		Level:   p.Level,
		Message: p.Notify,
	})
}
func (p *command_notify) SetNodes(nodes serverprotocol.Searchable) {
}
