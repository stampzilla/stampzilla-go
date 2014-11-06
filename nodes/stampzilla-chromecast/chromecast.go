package main

import (
	"fmt"
	"strings"
	"time"

	log "github.com/cihub/seelog"
	"github.com/davecgh/go-spew/spew"
	"github.com/jonaz/go-castv2"
	"github.com/jonaz/go-castv2/controllers"
	"github.com/jonaz/mdns"
)

type device struct {
	name string
}
type devices struct {
	devices []*device
}

func (d *devices) Add(name string) {
	newDevice := &device{name}
	d.devices = append(d.devices, newDevice)
}

func (d *devices) Get(name string) *device {
	for _, device := range d.devices {
		if name == device.name {
			return device
		}
	}
	return nil
}

type Chromecast struct {
	devices *devices
}

func NewChromecast() *Chromecast {
	c := &Chromecast{}
	c.devices = &devices{}
	return c
}

func (c *Chromecast) Listen() {
	entriesCh := make(chan *mdns.ServiceEntry, 4)
	go c.listen(entriesCh)
	go c.mdnsPeridicalFetcher(entriesCh)
}

func (c *Chromecast) listen(entriesCh chan *mdns.ServiceEntry) {
	// Make a channel for results and start listening
	for entry := range entriesCh {

		if !strings.Contains(entry.Name, "_googlecast._tcp") {
			return
		}

		if device := c.devices.Get(entry.Host); device != nil {
			return
		}

		c.devices.Add(entry.Host)

		fmt.Printf("Got new chromecast: %#v\n", entry)

		client, err := castv2.NewClient(entry.Addr, entry.Port)

		if err != nil {
			log.Error("Failed to connect to chromecast %s", entry.Addr)
		}

		//_ = controllers.NewHeartbeatController(client, "Tr@n$p0rt-0", "Tr@n$p0rt-0")

		heartbeat := controllers.NewHeartbeatController(client, "sender-0", "receiver-0")
		heartbeat.Start()

		connection := controllers.NewConnectionController(client, "sender-0", "receiver-0")
		connection.Connect()

		receiver := controllers.NewReceiverController(client, "sender-0", "receiver-0")
		response, err := receiver.GetStatus(time.Second * 5)

		spew.Dump("Status response", response, err)

		media := controllers.NewMediaController(client, "sender-0", "receiver-0")
		response, err = media.GetStatus(time.Second * 5)
		spew.Dump("Media response", response, err)
	}

}

func (c *Chromecast) mdnsPeridicalFetcher(entriesCh chan *mdns.ServiceEntry) {
	ticker := time.NewTicker(10 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				go func() {
					mdns.Query(&mdns.QueryParam{
						Service: "_googlecast._tcp",
						Domain:  "local",
						Timeout: time.Second * 5,
						Entries: entriesCh,
					})
				}()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

}
