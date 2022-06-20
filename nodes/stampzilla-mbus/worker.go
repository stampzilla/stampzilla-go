package main

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type Worker struct {
	wg     sync.WaitGroup
	work   chan func() error
	cancel context.CancelFunc
}

func NewWorker() *Worker {
	return &Worker{
		work: make(chan func() error),
	}
}

func (w *Worker) Do(fn func() error, defaultFn func() error) {
	select {
	case w.work <- fn:
	case <-time.After(time.Second):
		if defaultFn != nil {
			defaultFn()
		}
	}
}

func (w *Worker) Start(parentCtx context.Context, workers int) {
	var ctx context.Context
	ctx, w.cancel = context.WithCancel(parentCtx)
	w.wg.Add(workers)
	for i := 0; i < workers; i++ {
		go w.start(ctx)
	}
}

func (w *Worker) start(ctx context.Context) {
	defer w.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case fn := <-w.work:
			err := fn()
			if err != nil {
				logrus.Error(err)
			}
		}
	}
}
