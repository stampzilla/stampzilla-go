package main

import (
	"net"

	log "github.com/cihub/seelog"
	"github.com/stampzilla/gocast"
	"github.com/stampzilla/gocast/events"
)

type Chromecast struct {
	Id    string
	Name_ string `json:"Name"`

	PrimaryApp      string
	PrimaryEndpoint string
	//PlaybackActive  bool
	//Paused          bool

	IsStandBy     bool
	IsActiveInput bool

	Volume float64
	Muted  bool

	Addr net.IP
	Port int

	publish func()

	*gocast.Device
}

func NewChromecast(d *gocast.Device) *Chromecast {
	c := &Chromecast{
		Device: d,
	}

	d.OnEvent(c.Event)
	d.Connect()

	return c
}

func (c *Chromecast) Listen() {
}

func (c *Chromecast) Event(event events.Event) {
	switch data := event.(type) {
	case events.Connected:
		log.Info(c.Name(), "- Connected, weeihoo")

		c.Addr = c.Ip()
		c.Port = c.Device.Port()
		c.Id = c.Uuid()
		c.Name_ = c.Name()

		state.Add(c)
	case events.Disconnected:
		log.Warn(c.Name(), "- Disconnected, bah :/")

		state.Remove(c)

		c.Device.Connect()
	case events.AppStarted:
		c.PrimaryApp = data.DisplayName
		c.PrimaryEndpoint = data.TransportId

		log.Info(c.Name(), "- App started:", data.DisplayName, "(", data.AppID, ")")
	case events.AppStopped:
		c.PrimaryApp = ""
		log.Info(c.Name(), "- App stopped:", data.DisplayName, "(", data.AppID, ")")

	case events.ReceiverStatus:
		c.IsStandBy = data.Status.IsStandBy
		c.IsActiveInput = data.Status.IsActiveInput
		c.Volume = data.Status.Volume.Level
		c.Muted = data.Status.Volume.Muted

	//gocast.MediaEvent:
	default:
		log.Warn("unexpected event %T: %#v\n", data, data)
	}

	c.publish()
}
