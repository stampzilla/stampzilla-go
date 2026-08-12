package main

import (
	"context"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/v2/pkg/node"
)

func main() {
	args := os.Args[1:]

	// `stampzilla-dmx selftest ...` bypasses the server entirely to bring up
	// the DMX cable standalone - see selftest.go for why this has to be a
	// positional subcommand rather than a top-level flag.
	if selftestRequested(args) {
		os.Exit(runSelftestCommand(args[1:]))
	}
	if hint := legacySelftestHint(args); hint != "" {
		fmt.Fprintln(os.Stderr, hint)
		os.Exit(2)
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
