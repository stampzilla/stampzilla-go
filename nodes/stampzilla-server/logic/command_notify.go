package logic

import (
	log "github.com/cihub/seelog"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/notifications"
)

type command_notify struct {
	Notify             string                          `json:"notify"`
	Level              notifications.NotificationLevel `json:"level"`
	NotificationRouter notifications.Router            `json:"-" inject:""`
}

func (p *command_notify) Run(abort <-chan struct{}) {
	log.Infof("Running notification %#v", p)

	p.NotificationRouter.Dispatch(notifications.Notification{
		Level:   p.Level,
		Message: p.Notify,
	})
}
