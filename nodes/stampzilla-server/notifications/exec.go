package notifications

import (
	"os"
	"os/exec"
	"strings"
	"sync"

	log "github.com/cihub/seelog"
)

type Exec struct {
	wg        sync.WaitGroup
	Command   string
	Arguments []string
}

func (self *Exec) Start() {
}
func (self *Exec) Dispatch(note Notification) {
	self.wg.Add(1)
	defer self.wg.Done()

	args := make([]string, 0)
	for _, text := range self.Arguments {
		text = strings.Replace(text, "[level]", note.Level.String(), -1)
		text = strings.Replace(text, "[message]", note.Message, -1)
		text = strings.Replace(text, "[uuid]", note.SourceUuid, -1)
		text = strings.Replace(text, "[source]", note.Source, -1)

		args = append(args, text)
	}

	cmd := exec.Command(self.Command, args...)
	cmd.Env = os.Environ()

	out, err := cmd.Output()
	if err != nil {
		log.Error(err, " | ", out)
		return
	}

	log.Tracef("notifications.Exec result %s", out)
}
func (self *Exec) Stop() {
}
