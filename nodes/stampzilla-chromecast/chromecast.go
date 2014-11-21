package main

import (
	"strings"
	"time"

	"github.com/jonaz/mdns"
)

type Chromecast struct {
	Events  chan *Event `json:"-"`
	Devices *Devices
}

type Event struct {
	Name string
	Args []string
}

func NewChromecast() *Chromecast {
	c := &Chromecast{
		Events: make(chan *Event, 10),
	}
	c.Devices = &Devices{}
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
		// Only match chromecasts
		if !strings.Contains(entry.Name, "_googlecast._tcp") {
			continue
		}

		// Ignore devices that already are initiated
		if device := c.Devices.GetByName(entry.Host); device != nil {
			continue
		}

		// Add the new device
		c.Devices.AddMdnsEntry(entry, c.Events)
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
