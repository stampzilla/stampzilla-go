package logic

import "encoding/json"
import serverprotocol "github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/protocol"

type Actions interface {
	Run()
	GetByUuid(string) Action
}

type actions struct {
	Actions []*action
	nodes   *serverprotocol.Nodes
}

func (a *actions) Run() {
	for _, v := range a.Actions {
		v.Run()
	}
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
			action.SetNodes(a.nodes)
			for _, c := range action.Commands {
				c.nodes = a.nodes
			}
			a.Actions = append(a.Actions, action)
		}
		return
	}
	return
}
