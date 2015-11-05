package logic

import "encoding/json"
import serverprotocol "github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/protocol"

//type Actions interface {
//Run()
//GetByUuid(string) Action
//Actions() []Action
//}

type ActionService struct {
	Actions_ []*action             `json:"actions"`
	Nodes    *serverprotocol.Nodes `json:"-" inject:""`
}

func NewActions() *ActionService {
	return &ActionService{}
}

func (a *ActionService) Get() []*action {
	return a.Actions_
}
func (a *ActionService) Start() {
	a.Actions_ = make([]*action, 0)
	mapper := NewActionsMapper()
	mapper.Load(a)
}

func (a *ActionService) GetByUuid(uuid string) Action {
	for _, v := range a.Actions_ {
		if v.Uuid() == uuid {
			return v
		}

	}
	return nil
}

func (a *ActionService) UnmarshalJSON(b []byte) (err error) {
	type localActions []*action
	la := localActions{}
	if err = json.Unmarshal(b, &la); err == nil {
		for _, action := range la {
			for _, c := range action.Commands {
				c.SetNodes(a.Nodes)
			}
			a.Actions_ = append(a.Actions_, action)
		}
		return
	}
	return
}

func (a *ActionService) MarshalJSON() (res []byte, err error) {
	return json.Marshal(a.Actions_)
}
