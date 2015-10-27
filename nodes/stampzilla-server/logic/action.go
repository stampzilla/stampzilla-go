package logic

import "golang.org/x/net/context"

type Action interface {
	Run()
	Uuid() string
	Name() string
	Cancel()
}

type action struct {
	Name_    string     `json:"name"`
	Uuid_    string     `json:"uuid"`
	Commands []*command `json:"commands"`
	cancel   context.CancelFunc
}

func (a *action) Uuid() string {
	return a.Uuid_
}
func (a *action) Name() string {
	return a.Name_
}
func (a *action) Cancel() {
	if a.cancel != nil {
		a.cancel()
	}
}
func (a *action) Run() {
	a.Cancel()
	a.run()
}
func (a *action) run() {
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
