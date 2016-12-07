package main

import (
	"fmt"
	"net"
	"time"

	log "github.com/cihub/seelog"
	"github.com/stampzilla/gocast"
	"github.com/stampzilla/gocast/events"
	"github.com/stampzilla/gocast/handlers"
	"github.com/stampzilla/gocast/responses"
)

type Chromecast struct {
	Id    string
	Name_ string `json:"Name"`

	PrimaryApp      string
	PrimaryEndpoint string
	Playing         bool
	//Paused          bool

	IsStandBy     bool
	IsIdleScreen  bool
	IsActiveInput bool

	Volume float64
	Muted  bool

	Addr net.IP
	Port int

	Media Chromecast_Media

	publish func()

	mediaHandler           *handlers.Media
	mediaConnectionHandler *handlers.Connection

	appLaunch chan string

	*gocast.Device
}

type Chromecast_Media struct {
	Title    string
	SubTitle string
	Url      string
	Thumb    string
	Duration float64
}

func NewChromecast(d *gocast.Device) *Chromecast {
	c := &Chromecast{
		Device:                 d,
		mediaHandler:           &handlers.Media{},
		mediaConnectionHandler: &handlers.Connection{},
		appLaunch:              make(chan string),
	}

	c.OnEvent(c.Event)
	return c
}

func (c *Chromecast) Play() {
	c.mediaHandler.Play()
}
func (c *Chromecast) Pause() {
	c.mediaHandler.Pause()
}
func (c *Chromecast) Stop() {
	c.mediaHandler.Stop()
}

func (c *Chromecast) PlayUrl(url string, contentType string) {
	err := c.Device.ReceiverHandler.LaunchApp(gocast.AppMedia)
	if err != nil && err != handlers.ErrAppAlreadyLaunched {
		log.Error(err)
		return
	}

	if err != handlers.ErrAppAlreadyLaunched {
		//Wait for new media connection to launched app
		if err := c.waitForAppLaunch(gocast.AppMedia); err != nil {
			log.Error(err)
			return
		}
	}

	if contentType == "" {
		contentType = "audio/mpeg"
	}
	item := responses.MediaItem{
		ContentId:   url,
		StreamType:  "BUFFERED",
		ContentType: contentType,
	}
	err = c.mediaHandler.LoadMedia(item, 0, true, map[string]interface{}{})
	if err != nil {
		log.Error(err)
		return
	}
}

func (c *Chromecast) waitForAppLaunch(app string) error {
	select {
	case launchedApp := <-c.appLaunch:
		if app == launchedApp {
			return nil
		}
		return fmt.Errorf("Wrong app launched. Expected %s got %s", app, launchedApp)
	case <-time.After(time.Second * 20):
		return fmt.Errorf("timeout waiting for app launch after 20 seconds")

	}

}
func (c *Chromecast) appLaunched(app string) {
	select {
	case c.appLaunch <- app:
		log.Info("Notified c.appLaunch")
	case <-time.After(time.Second * 2):
		log.Info("No one is waiting for appLaunch event")
	}
}

func (c *Chromecast) Event(event events.Event) {
	switch data := event.(type) {
	case events.Connected:
		log.Info(c.Name(), "- Connected")

		c.Addr = c.Ip()
		c.Port = c.Device.Port()
		c.Id = c.Uuid()
		c.Name_ = c.Name()

		state.Add(c)
	case events.Disconnected:
		log.Warn(c.Name(), "- Disconnected")

		state.Remove(c)
	case events.AppStarted:
		log.Info(c.Name(), "- App started:", data.DisplayName, "(", data.AppID, ")")
		//spew.Dump("Data:", data)

		c.PrimaryApp = data.DisplayName
		c.PrimaryEndpoint = data.TransportId
		c.IsIdleScreen = data.IsIdleScreen

		c.Media.Title = data.DisplayName
		if c.IsIdleScreen {
			c.Media.Title = ""
		}
		c.Media.SubTitle = ""
		c.Media.Thumb = ""
		c.Media.Url = ""
		c.Media.Duration = 0

		//If the app supports media controls lets subscribe to it
		if data.HasNamespace("urn:x-cast:com.google.cast.media") {
			log.Info(c.Name(), "- Subscribe cast.tp.connection:", data.DisplayName, "(", data.AppID, ")")
			c.Subscribe("urn:x-cast:com.google.cast.tp.connection", data.TransportId, c.mediaConnectionHandler)
			log.Info(c.Name(), "- Subscribe cast.media:", data.DisplayName, "(", data.AppID, ")")
			c.Subscribe("urn:x-cast:com.google.cast.media", data.TransportId, c.mediaHandler)
		}
		c.appLaunched(data.AppID)
		log.Info(c.Name(), "- Notifying appLanunched:", data.DisplayName, "(", data.AppID, ")")

	case events.AppStopped:
		log.Info(c.Name(), "- App stopped:", data.DisplayName, "(", data.AppID, ")")
		//spew.Dump("Data:", data)

		//unsubscribe from old channels
		for _, v := range data.Namespaces {
			c.UnsubscribeByUrnAndDestinationId(v.Name, data.TransportId)
		}
		c.PrimaryApp = ""
		c.PrimaryEndpoint = ""
		c.Playing = false

		c.Media.Title = ""
		c.Media.SubTitle = ""
		c.Media.Thumb = ""
		c.Media.Url = ""
		c.Media.Duration = 0

	case events.ReceiverStatus:
		c.IsStandBy = data.Status.IsStandBy
		c.IsActiveInput = data.Status.IsActiveInput
		c.Volume = data.Status.Volume.Level
		c.Muted = data.Status.Volume.Muted
	case events.Media:
		playing := c.Playing
		if data.PlayerState == "PLAYING" {
			c.Playing = true
		} else {
			c.Playing = false
		}

		if data.PlayerState == "IDLE" {
			c.Media.Title = c.PrimaryApp
			c.Media.SubTitle = ""
			c.Media.Thumb = ""
			c.Media.Url = ""
			c.Media.Duration = 0
		}

		if data.Media != nil {
			c.Media.Title = c.PrimaryApp
			if data.Media.MetaData.Title != "" {
				c.Media.Title = data.Media.MetaData.Title
			}
			if c.IsIdleScreen {
				c.Media.Title = ""
			}

			c.Media.SubTitle = data.Media.MetaData.SubTitle
			c.Media.Url = data.Media.ContentId
			c.Media.Duration = data.Media.Duration
			if len(data.Media.MetaData.Images) > 0 {
				c.Media.Thumb = data.Media.MetaData.Images[0].Url
			} else {
				c.Media.Thumb = ""
			}
		}

		//Only publish if playing state changed
		if playing != c.Playing {
			c.publish()
		}
		return

	//gocast.MediaEvent:
	default:
		log.Warn("unexpected event %T: %#v\n", data, data)
	}

	c.publish()
}
