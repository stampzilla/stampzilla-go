package main

import (
	"flag"
	"fmt"
	"net/http"

	log "github.com/cihub/seelog"
	evdev "github.com/gvalkov/golang-evdev"
	"github.com/stampzilla/stampzilla-go/nodes/basenode"
	"github.com/stampzilla/stampzilla-go/pkg/notifier"
)

var notify *notifier.Notify

type Config struct {
	Device string            `json:"device"`
	Keys   map[string]string `json:"keys"`
}

func main() {
	log.Info("Starting hidcommander node")

	// Parse all commandline arguments, host and port parameters are added in the basenode init function
	flag.Parse()

	//Get a config with the correct parameters
	config := basenode.NewConfig()

	//Activate the config
	basenode.SetConfig(config)

	nc := &Config{}
	err := config.NodeSpecific(&nc)
	if err != nil {
		log.Error(err)
		return
	}

	device, err := evdev.Open(nc.Device)
	if err != nil {
		log.Error(err)
		return
	}

	var event *evdev.InputEvent
	for {
		event, err = device.ReadOne()
		key := evdev.NewKeyEvent(event)
		if key.State != evdev.KeyDown {
			continue
		}
		for k, v := range nc.Keys {
			if k == evdev.KEY[int(key.Scancode)] {
				getRequest(config, v)
			}
		}
	}
	//select {}
}

func getRequest(config *basenode.Config, cmd string) {
	url := "http://" + config.Host + ":" + config.Port + "/api/" + cmd
	fmt.Println(url)
	_, err := http.Get(url)
	if err != nil {
		log.Error(err)
	}
}
