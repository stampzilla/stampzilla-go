package main

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/v2/pkg/node"
)

func main() {
	// -selftest-port bypasses the server entirely to bring up the DMX cable
	// standalone - see selftest.go.
	if handled, code := maybeRunSelftest(os.Args[1:]); handled {
		os.Exit(code)
	}

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
