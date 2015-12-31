package logic

import (
	"time"

	serverprotocol "github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/protocol"
)

type notify struct {
	Notify string `json:"notify"`
	notify time.Duration
	nodes  serverprotocol.Searchable
}

func NewNotify(duration string) *notify {
	p := &notify{}
	p.SetDuration(duration)
	return p
}

func (p *notify) Uuid() string {
	return ""
}
func (p *notify) Run() {
	<-time.After(p.notify)
}
func (p *notify) SetNodes(nodes serverprotocol.Searchable) {
	p.nodes = nodes
}

func (p *notify) SetDuration(duration string) error {
	d, err := time.ParseDuration(duration)

	if err != nil {
		return err
	}

	p.Notify = duration
	p.notify = d
	return nil
}
