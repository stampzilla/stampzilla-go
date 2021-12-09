package main

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/gocast/discovery"
	"github.com/stampzilla/stampzilla-go/v2/pkg/node"
)

var state = &State{
	Chromecasts: make(map[string]*Chromecast),
}

func main() {
	node := node.New("chromecast")

	// node.OnConfig(updatedConfig)

	ctx, cancel := context.WithCancel(context.Background())
	node.OnShutdown(func() {
		cancel()
	})

	if err := node.Connect(); err != nil {
		logrus.Error(err)
		return
	}

	discovery := discovery.NewService()
	go discoveryListner(ctx, node, discovery)
	discovery.Start(ctx, time.Second*10)

	node.Wait()
}

func discoveryListner(ctx context.Context, node *node.Node, discovery *discovery.Service) {
	for {
		select {
		case device := <-discovery.Found():
			logrus.Debugf("New device discovered: %s", device.String())
			d := NewChromecast(node, device)
			state.Add(d)
			err := device.Connect(ctx)
			if err != nil {
				logrus.Error(err)
			}
		case <-ctx.Done():
			return
		}
	}
}
