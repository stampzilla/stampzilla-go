package logic

import "encoding/json"
import serverprotocol "github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/protocol"

type Actions interface {
	Run()
	GetByUuid(string) Action
}

type actions struct {
	Actions []*action
	Nodes   *serverprotocol.Nodes `inject:""`
}

func NewActions() *actions {
	return &actions{}
}

func (a *actions) Run() {
	for _, v := range a.Actions {
		v.Run()
	}
}
func (a *actions) Start() {
	mapper := newActionsMapper()
	mapper.Load(a)
}

func (a *actions) GetByUuid(uuid string) Action {
	for _, v := range a.Actions {
		if v.Uuid() == uuid {
			return v
		}

	}
	return nil
}

func (a *actions) UnmarshalJSON(b []byte) (err error) {
	type localActions actions
	la := localActions{}
	if err = json.Unmarshal(b, &la); err == nil {
		for _, action := range la.Actions {
			action.SetNodes(a.Nodes)
			for _, c := range action.Commands {
				c.nodes = a.Nodes
			}
			a.Actions = append(a.Actions, action)
		}
		return
	}
	return
}
