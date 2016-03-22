package logic

import (
	"encoding/json"

	log "github.com/cihub/seelog"
	"golang.org/x/net/context"
)

type Action interface {
	Run()
	Uuid() string
	Name() string
	Cancel()
}

type action struct {
	Name_    string    `json:"name"`
	Uuid_    string    `json:"uuid"`
	Commands []Command `json:"commands"`
	cancel   context.CancelFunc
}

func (a *action) Uuid() string {
	return a.Uuid_
}
func (a *action) Name() string {
	return a.Name_
}
func (a *action) Cancel() {
	log.Debugf("Cancel action %s", a.Uuid())
	if a.cancel != nil {
		a.cancel()
	}
}
func (a *action) Run() {
	a.Cancel()
	a.run()
}
func (a *action) run() {
	log.Debugf("Running action %s", a.Uuid())
	ctx, cancel := context.WithCancel(context.Background())
	a.cancel = cancel
	queue := make(chan Command)
	go func() {
		for {
			select {
			case cmd := <-queue:
				cmd.Run()
			case <-ctx.Done():
				a.cancel = nil
				return
			}
		}
	}()

	go func() {
		for _, cmd := range a.Commands {
			select {
			case queue <- cmd:
			case <-ctx.Done():
				return

			}
		}
		cancel()
	}()
}

func (a *action) UnmarshalJSON(b []byte) (err error) {

	type localAction struct {
		Name_    string            `json:"name"`
		Uuid_    string            `json:"uuid"`
		Commands []json.RawMessage `json:"commands"`
	}

	la := localAction{}
	if err = json.Unmarshal(b, &la); err == nil {
		for _, action := range la.Commands {
			cmd, err := a.unmarshalJSONcommands(action)
			if err != nil {
				return err
			}
			a.Commands = append(a.Commands, cmd)
		}
	}

	a.Name_ = la.Name_
	a.Uuid_ = la.Uuid_

	return
}

func (a *action) unmarshalJSONcommands(b []byte) (cmd Command, err error) {
	test := make(map[string]interface{})
	if err = json.Unmarshal(b, &test); err != nil {
		return nil, err
	}

	if _, ok := test["pause"]; ok {
		cmd = &command_pause{}
	} else if _, ok := test["notify"]; ok {
		cmd = &command_notify{}
	} else {
		cmd = &command{}
	}

	err = json.Unmarshal(b, cmd)
	return cmd, err
}
