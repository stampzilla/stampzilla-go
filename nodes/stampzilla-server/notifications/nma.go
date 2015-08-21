package notifications

import (
	log "github.com/cihub/seelog"
	"github.com/dustin/go-nma"
)

// NMA - NotifyMyAndroid
type Nma struct {
	nma    *nma.NMA
	ApiKey string
}

func (self *Nma) Start() {
	self.nma = nma.New(self.ApiKey)
}
func (self *Nma) Dispatch(note Notification) {
	log.Trace("Nma deliver start")
	e := nma.Notification{
		Event:       note.Message,
		Description: note.Source,
		Priority:    1,
	}

	if err := self.nma.Notify(&e); err != nil {
		log.Error("Error sending message:  %v\n", err)
	}

	log.Trace("Nma deliver done")

}
func (self *Nma) Stop() {
}
