package main

import (
	"net"
	"time"

	"code.google.com/p/go-uuid/uuid"

	log "github.com/cihub/seelog"
	"github.com/davecgh/go-spew/spew"
	"github.com/jonaz/go-castv2"
	"github.com/jonaz/go-castv2/controllers"
	"github.com/jonaz/mdns"
)

type Device struct {
	Id string

	PrimaryApp      string
	PrimaryEndpoint string
	PlaybackActive  bool
	Paused          bool

	IsStandBy     bool
	IsActiveInput bool

	Volume float64
	Muted  bool

	activeSubscriptions map[string]interface{}

	Name string
	Addr net.IP
	Port int

	events chan *Event

	client *castv2.Client
}

type Devices struct {
	Devices map[string]*Device
}

func (c *Devices) GetState() interface{} {
	return &c.Devices
}

func (d *Device) Changed() {
	d.Fire("Updated")
}

func (d *Device) Fire(name string, args ...string) {
	select {
	case d.events <- &Event{
		Name: name,
		Args: args,
	}:
	default:
		log.Warn("Lost event: ", name, " Args: ", args)
	}
}

func (d *Devices) AddMdnsEntry(entry *mdns.ServiceEntry, events chan *Event) {
	newDevice := &Device{
		Id:     uuid.New(),
		Name:   entry.Host,
		Addr:   entry.Addr,
		Port:   entry.Port,
		events: events,

		activeSubscriptions: make(map[string]interface{}, 0),
	}

	go newDevice.Worker()

	if d.Devices == nil {
		d.Devices = make(map[string]*Device)
	}

	d.Devices[newDevice.Id] = newDevice

	newDevice.Fire("Added", newDevice.Id)
}

func (d *Devices) Get(name string) *Device {
	for _, device := range d.Devices {
		if name == device.Id {
			return device
		}
	}
	return nil
}

func (d *Devices) GetByName(name string) *Device {
	for _, device := range d.Devices {
		if name == device.Name {
			return device
		}
	}
	return nil
}

func (d *Device) Worker() {
	log.Infof("Got new chromecast: %#v\n", d.Name)

	var err error
	d.client, err = castv2.NewClient(d.Addr, d.Port)

	if err != nil {
		log.Error("Failed to connect to chromecast %s", d.Name)
	}

	receiver := controllers.NewReceiverController(d.client, "sender-0", "receiver-0")
	go func() {
		for {
			select {
			case msg := <-receiver.Incoming:
				if msg == nil {
					log.Warn("ReciverController shutdown")
					return
				}

				//spew.Dump("Status response", msg)

				d.IsStandBy = msg.Status.IsStandBy
				d.IsActiveInput = msg.Status.IsActiveInput

				d.Volume = *msg.Status.Volume.Level
				d.Muted = *msg.Status.Volume.Muted

				d.PrimaryApp = ""
				d.PrimaryEndpoint = ""

				for _, app := range msg.Status.Applications {
					if d.PrimaryApp == "" {
						d.PrimaryApp = app.DisplayName
						d.PrimaryEndpoint = app.TransportId
					}

					for _, namespace := range app.Namespaces {
						// Skip already activated subscriptions
						if _, ok := d.activeSubscriptions[namespace.Name]; ok {
							continue
						}

						switch namespace.Name {
						case "urn:x-cast:com.google.cast.media":
							d.activeSubscriptions[namespace.Name] =
								d.mediaSubscriber(app.TransportId, namespace.Name)
						case "urn:x-cast:plex":
							d.activeSubscriptions[namespace.Name] =
								d.plexSubscriber(app.TransportId, namespace.Name)
						}
					}
				}

				d.Changed()
			}
		}
	}()

	receiver.GetStatus(time.Second * 1)
}

func (d *Device) mediaSubscriber(transportId, namespace string) *controllers.MediaController {
	media := controllers.NewMediaController(d.client, "receiver-0", transportId)

	go func() {
		for {
			msg := <-media.Incoming
			if media.Incoming == nil {
				log.Info("Disconnected, removing from active subscriptions")
				delete(d.activeSubscriptions, namespace)
				// Disconnected
				return
			}

			for _, state := range msg.Status {
				switch state.PlayerState {
				case "PLAYING":
					d.PlaybackActive = true
					d.Paused = false
				case "PAUSED":
					d.PlaybackActive = true
					d.Paused = true
				default:
					d.PlaybackActive = false
					d.Paused = false
				}
			}

			d.Changed()

			//spew.Dump("Media response", msg)
		}
	}()

	return media
}

func (d *Device) plexSubscriber(transportId, namespace string) *controllers.MediaController {
	media := controllers.NewPlexController(d.client, "receiver-0", transportId)

	go func() {
		for {
			msg := <-media.Incoming
			if media.Incoming == nil {
				log.Info("Disconnected, removing from active subscriptions")
				delete(d.activeSubscriptions, namespace)
				// Disconnected
				return
			}

			spew.Dump("Media response", msg)
		}
	}()

	return media
}
