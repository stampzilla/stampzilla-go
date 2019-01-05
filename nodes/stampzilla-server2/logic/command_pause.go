package logic

import (
	"encoding/json"
	"time"

	"github.com/sirupsen/logrus"
)

type commandPause struct {
	Pause string `json:"pause"`
	pause time.Duration
}

func NewPause(duration string) *commandPause {
	p := &commandPause{}
	p.SetDuration(duration)
	return p
}

func (p *commandPause) Run(abort <-chan struct{}) {
	logrus.Infof("Pausing for %s", p.Pause)
	select {
	case <-time.After(p.pause):
	case <-abort:
	}
}
func (p *commandPause) SetDuration(duration string) error {
	d, err := time.ParseDuration(duration)

	if err != nil {
		return err
	}

	p.Pause = duration
	p.pause = d
	return nil
}

func (p *commandPause) UnmarshalJSON(b []byte) (err error) {
	type localCmd struct {
		Pause string `json:"pause"`
	}
	cmd := localCmd{}
	err = json.Unmarshal(b, &cmd)
	p.SetDuration(cmd.Pause)
	p.Pause = cmd.Pause
	return
}
