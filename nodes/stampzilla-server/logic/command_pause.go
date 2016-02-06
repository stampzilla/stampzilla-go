package logic

import (
	"encoding/json"
	"time"

	serverprotocol "github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/protocol"
)

type command_pause struct {
	Pause string `json:"pause"`
	pause time.Duration
}

func NewPause(duration string) *command_pause {
	p := &command_pause{}
	p.SetDuration(duration)
	return p
}

func (p *command_pause) Run() {
	<-time.After(p.pause)
}
func (p *command_pause) SetNodes(nodes serverprotocol.Searchable) {
}

func (p *command_pause) SetDuration(duration string) error {
	d, err := time.ParseDuration(duration)

	if err != nil {
		return err
	}

	p.Pause = duration
	p.pause = d
	return nil
}

func (p *command_pause) UnmarshalJSON(b []byte) (err error) {
	type localCmd struct {
		Pause string `json:"pause"`
	}
	cmd := localCmd{}
	err = json.Unmarshal(b, &cmd)
	p.SetDuration(cmd.Pause)
	p.Pause = cmd.Pause
	return
}
