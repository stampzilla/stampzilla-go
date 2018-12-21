package logic

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
)

type Action interface {
	Run(chan ActionProgress)
	Uuid() string
	Name() string
	Cancel()
}

type ActionProgress struct {
	Address string `json:"address"`
	Uuid    string `json:"uuid"`
	Step    int    `json:"step"`
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
	logrus.Debugf("Cancel action %s", a.Uuid())
	if a.cancel != nil {
		a.cancel()
	}
}
func (a *action) Run(c chan ActionProgress) {
	a.Cancel()
	a.run(c)
}
func (a *action) run(progressChan chan ActionProgress) {
	logrus.Debugf("Running action %s", a.Uuid())
	ctx, cancel := context.WithCancel(context.Background())
	a.cancel = cancel
	queue := make(chan Command)

	addr := 0 // To generate a unique pointer address

	go func() {
		for {
			select {
			case <-ctx.Done():
				//a.cancel = nil // this was causing it to miss cancel sometimes
				a.tryNotifyProgress(&addr, progressChan, -1) // Done
				return
			case cmd := <-queue:
				cmd.Run(ctx.Done())
			}
		}
	}()

	go func() {

		a.tryNotifyProgress(&addr, progressChan, 0) // Start
		for index, cmd := range a.Commands {
			select {
			case <-ctx.Done():
				return
			case queue <- cmd:
				a.tryNotifyProgress(&addr, progressChan, index+1) // Notify start of step n
			}
		}
		a.tryNotifyProgress(&addr, progressChan, -1) // Done
		cancel()
	}()
}

func (a *action) tryNotifyProgress(addr *int, c chan ActionProgress, step int) {
	msg := ActionProgress{
		Address: fmt.Sprintf("%p", addr),
		Uuid:    a.Uuid(),
		Step:    step,
	}

	// Try to deliver message
	select {
	case c <- msg:
	default:
		logrus.Warnf("Dropped progress notification for action runner %p", addr)
	}
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
