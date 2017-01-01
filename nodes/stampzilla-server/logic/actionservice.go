package logic

import (
	"encoding/json"
	"sync"

	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/notifications"
	serverprotocol "github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/protocol"
)

//type Actions interface {
//Run()
//GetByUuid(string) Action
//Actions() []Action
//}

type ActionService struct {
	actions            []Action              `json:"actions"`
	Nodes              *serverprotocol.Nodes `json:"-" inject:""`
	NotificationRouter notifications.Router  `json:"-" inject:""`
	sync.RWMutex
}

func NewActionService() *ActionService {
	return &ActionService{}
}

func (as *ActionService) Add(a Action) {
	as.Lock()
	//r.actions = append(r.actions, a)
	as.actions = append(as.actions, a)
	as.Unlock()
}

func (a *ActionService) Get() []Action {
	a.RLock()
	defer a.RUnlock()
	return a.actions
}
func (a *ActionService) Start() {
	a.actions = make([]Action, 0)
	mapper := NewActionsMapper()
	mapper.Load(a)
}

func (a *ActionService) GetByUuid(uuid string) Action {
	for _, v := range a.actions {
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
				a.SetCommandDependencies(c)
			}
			a.actions = append(a.actions, action)
		}
		return
	}
	return
}
func (a *ActionService) SetCommandDependencies(cmd Command) {
	switch c := cmd.(type) {
	case *command:
		c.nodes = a.Nodes
	case *command_notify:
		c.NotificationRouter = a.NotificationRouter
	}
}

func (a *ActionService) MarshalJSON() (res []byte, err error) {
	return json.Marshal(a.actions)
}
