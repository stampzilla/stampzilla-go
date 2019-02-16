package main

import (
	"time"

	"github.com/sirupsen/logrus"

	"github.com/stampzilla/gocast/discovery"
	"github.com/stampzilla/stampzilla-go/pkg/node"
)

var state = &State{
	Chromecasts: make(map[string]*Chromecast),
}

func main() {
	node := node.New("chromecast")

	//node.OnConfig(updatedConfig)

	stop := make(chan struct{})
	node.OnShutdown(func() {
		close(stop)
	})

	err := node.Connect()
	if err != nil {
		logrus.Error(err)
		return
	}

	discovery := discovery.NewService()
	discovery.Periodic(time.Second * 10)
	go discoveryListner(node, discovery, stop)

	node.Wait()
}

func discoveryListner(node *node.Node, discovery *discovery.Service, stop chan struct{}) {
	for {
		select {
		case device := <-discovery.Found():
			logrus.Debugf("New device discoverd: %s", device.String())
			d := NewChromecast(node, device)
			state.Add(d)
			go func() {
				err := device.Connect()
				if err != nil {
					logrus.Error(err)
				}
			}()
		case <-stop:
			return
		}
	}
}

// THis is called on each incomming command
/*
func processCommand(cmd protocol.Command) {
	logrus.Info("Incoming command from server:", cmd)
	if len(cmd.Args) == 0 {
		logrus.Error("Missing argument 0 (which player?)")
		return
	}
	player := state.GetByUUID(cmd.Args[0])
	if player == nil {
		logrus.Errorf("Player with id %s not found", cmd.Args[0])
		return
	}

	switch cmd.Cmd {
	case "play":
		if len(cmd.Args) > 1 && strings.HasPrefix(cmd.Args[1], "http") {
			player.PlayUrl(strings.Join(cmd.Args[1:], "/"), "")
			return
		}
		player.Play()
	case "pause":
		player.Pause()
	case "stop":
		player.Stop()
	}
}
*/
