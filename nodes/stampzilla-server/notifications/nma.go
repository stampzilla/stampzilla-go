package notifications

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"

	log "github.com/cihub/seelog"
)

// NMA - NotifyMyAndroid
type Nma struct {
	wg     sync.WaitGroup
	ApiKey string
}

func (self *Nma) Start() {
}
func (self *Nma) Dispatch(note Notification) {
	self.wg.Add(1)
	defer self.wg.Done()

	resp, err := http.PostForm("https://www.notifymyandroid.com/publicapi/notify", url.Values{"apikey": {self.ApiKey}, "application": {"Stampzilla"}, "event": {note.Message}, "description": {note.Source}})
	if err != nil {
		log.Warn("Failed to deliver notification: ", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	log.Trace("Nma deliver result: ", string(body))

}
func (self *Nma) Stop() {
}
