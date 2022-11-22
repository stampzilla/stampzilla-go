package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"

	"github.com/jonaz/gombus"
	"github.com/sirupsen/logrus"
)

type Worker struct {
	wg        sync.WaitGroup
	work      chan func(*gombus.Conn) error
	reconnect chan struct{}
	cancel    context.CancelFunc
	conn      *gombus.Conn
	config    *Config
}

func NewWorker(config *Config) *Worker {
	return &Worker{
		work:      make(chan func(*gombus.Conn) error),
		reconnect: make(chan struct{}),
		config:    config,
	}
}

func (w *Worker) Do(fn func(*gombus.Conn) error) {
	w.work <- fn
}

func (w *Worker) Start(parentCtx context.Context, workers int) {
	var ctx context.Context
	ctx, w.cancel = context.WithCancel(parentCtx)

	go func() {
		defer w.wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case <-w.reconnect:
				for {
					err := w.connectTCP()
					if err != nil {
						logrus.Error(err)
						time.Sleep(time.Second * 1)
						continue
					}
					break
				}
			}
		}
	}()
	w.reconnect <- struct{}{}
	for i := 0; i < workers; i++ {
		w.wg.Add(1)
		go w.start(ctx)
	}
}

func (w *Worker) connectTCP() error {
	var err error
	w.conn, err = gombus.Dial(net.JoinHostPort(w.config.Host, w.config.Port))
	if err != nil {
		return fmt.Errorf("error connecting to mbus: %w", err)
	}

	if conn, ok := w.conn.Conn().(*net.TCPConn); ok {
		er := conn.SetKeepAlive(true)
		if er != nil {
			return er
		}

		er = conn.SetKeepAlivePeriod(time.Second * 10)
		if er != nil {
			return er
		}
	}

	return nil
}

func (w *Worker) start(ctx context.Context) {
	defer w.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case fn := <-w.work:
			err := fn(w.conn)
			if err != nil {
				logrus.Error(err)
			}

			if os.IsTimeout(err) || errors.Is(err, io.EOF) {
				w.reconnect <- struct{}{}
			}
		}
	}
}
