package logic

import (
	"time"

	serverprotocol "github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/protocol"
)

type pause struct {
	Pause string `json:"pause"`
	pause time.Duration
}

func NewPause(duration string) *pause {
	p := &pause{}
	p.SetDuration(duration)
	return p
}

func (p *pause) Uuid() string {
	return ""
}
func (p *pause) Run() {
	<-time.After(p.pause)
}
func (p *pause) SetNodes(nodes serverprotocol.Searchable) {
}

func (p *pause) SetDuration(duration string) error {
	d, err := time.ParseDuration(duration)

	if err != nil {
		return err
	}

	p.Pause = duration
	p.pause = d
	return nil
}
