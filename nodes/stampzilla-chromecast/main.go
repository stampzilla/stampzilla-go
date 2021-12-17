package main

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/gocast/discovery"
	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/models/devices"
	"github.com/stampzilla/stampzilla-go/v2/pkg/node"
)

var chromecasts = &State{
	Chromecasts: make(map[string]*Chromecast),
}

func main() {
	node := node.New("chromecast")

	// node.OnConfig(updatedConfig)

	ctx, cancel := context.WithCancel(context.Background())
	node.OnShutdown(func() {
		cancel()
	})
	node.OnRequestStateChange(stateChange)

	if err := node.Connect(); err != nil {
		logrus.Error(err)
		return
	}

	discovery := discovery.NewService()
	go discoveryListner(ctx, node, discovery)
	discovery.Start(ctx, time.Second*10)

	node.Wait()
}
func stateChange(state devices.State, device *devices.Device) error {
	var err error
	state.String("say", func(text string) {
		cc := chromecasts.GetByUUID(device.ID.ID)
		if cc != nil {
			base := "https://translate.google.com/translate_tts?client=tw-ob&ie=UTF-8&q=%s&tl=%s"
			u := fmt.Sprintf(base, url.QueryEscape(text), url.QueryEscape("sv"))
			cc.PlayURL(u, "audio/mpeg")
		}
	})
	if err != nil {
		return err
	}

	return err
}

func discoveryListner(ctx context.Context, node *node.Node, discovery *discovery.Service) {
	for {
		select {
		case device := <-discovery.Found():
			logrus.Debugf("New device discovered: %s", device.String())
			d := NewChromecast(node, device)
			chromecasts.Add(d)
			err := device.Connect(ctx)
			if err != nil {
				logrus.Error(err)
			}
		case <-ctx.Done():
			return
		}
	}
}
