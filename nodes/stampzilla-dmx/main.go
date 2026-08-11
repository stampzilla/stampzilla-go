package main

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/v2/pkg/node"
)

func main() {
	n, e := start()
	if n == nil {
		return
	}
	n.Wait()
	e.close()
}

func start() (*node.Node, *engine) {
	n := node.New("dmx")
	e := newEngine(n)

	n.OnConfig(e.updatedConfig)

	if err := n.Connect(); err != nil {
		logrus.Error(err)
		return nil, nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	n.OnShutdown(cancel)

	n.OnRequestStateChange(e.onRequestStateChange)

	go e.run(ctx)

	return n, e
}
