package notifications

import (
	"sync"

	log "github.com/cihub/seelog"
	"github.com/dustin/go-nma"
)

// NMA - NotifyMyAndroid
type Nma struct {
	wg     sync.WaitGroup
	nma    *nma.NMA
	ApiKey string
}

func (self *Nma) Start() {
	self.nma = nma.New(self.ApiKey)
}
func (self *Nma) Dispatch(note Notification) {
	self.wg.Add(1)
	defer self.wg.Done()

	log.Trace("Nma deliver start")
	e := nma.Notification{
		Event:       note.Message,
		Description: note.Source,
		Priority:    1,
	}

	if err := self.nma.Notify(&e); err != nil {
		log.Error("Error sending message:  %v\n", err)
	}

	//resp, err := http.PostForm("https://www.notifymyandroid.com/publicapi/notify", url.Values{"apikey": {self.ApiKey}, "application": {"Stampzilla"}, "event": {note.Message}, "description": {note.Source}})
	//if err != nil {
	//log.Warn("Failed to deliver notification: ", err)
	//}
	//defer resp.Body.Close()
	//body, err := ioutil.ReadAll(resp.Body)

	log.Trace("Nma deliver done")

}
func (self *Nma) Stop() {
}
