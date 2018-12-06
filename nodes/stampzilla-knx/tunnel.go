package main

import (
	"log"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/models"
	"github.com/stampzilla/stampzilla-go/pkg/node"
	"github.com/vapourismo/knx-go/knx"
	"github.com/vapourismo/knx-go/knx/dpt"
)

type tunnel struct {
	Node    *node.Node
	Address string
	Client  knx.GroupTunnel

	Groups map[string][]groupLink
	sync.RWMutex
}

type groupLink struct {
	Name   string
	Type   string
	Device *models.Device
}

func newTunnel(node *node.Node) *tunnel {
	return &tunnel{
		Node:   node,
		Groups: make(map[string][]groupLink),
	}
}

func (tunnel *tunnel) SetAddress(address string) {
	if address == tunnel.Address {
		return
	}

	client, err := knx.NewGroupTunnel(address, knx.DefaultTunnelConfig)
	if err != nil {
		log.Fatal(err)
	}

	if tunnel.Client.Tunnel != nil {
		tunnel.Client.Close()
	}

	go func() {
		tunnel.Address = address
		tunnel.Client = client
		defer client.Close()

		for msg := range client.Inbound() {
			logrus.Warnf("Got message %+v", msg)
			tunnel.RLock()
			if links, ok := tunnel.Groups[msg.Destination.String()]; ok {
				for _, gl := range links {
					logrus.Info("Found link", gl.Name, gl.Device.ID)

					var value interface{}
					var err error
					switch gl.Type {
					case "bool":
						value = new(dpt.Switch)
					case "temperature":
						value = new(dpt.ValueTemp) //2 bytes floating point
					case "lux":
						value = new(dpt.ValueTemp) //2 bytes floating point
					}

					if dptv, ok := value.(dpt.DatapointValue); ok {
						err = dptv.Unpack(msg.Data)
						if err != nil {
							logrus.Error("Failed to unpack", err)
							continue
						}

						gl.Device.State[gl.Name] = dptv
						tunnel.Node.AddOrUpdate(gl.Device)
					} else {
						logrus.Warn("Unsupported type %s", gl.Type)
					}
				}
			}
			tunnel.RUnlock()

			//var temp dpt.ValueTemp

			//err := temp.Unpack(msg.Data)
			//if err != nil {
			//util.Logger.Printf("%+v", msg)
			//continue
			//}

			//util.Logger.Printf("%+v: %v", msg, temp)
		}
	}()
}

func (tunnel *tunnel) AddLink(ga string, n string, t string, d *models.Device) {
	logrus.Warnf("Add link for %s to %s", ga, d.ID)
	tunnel.Lock()
	if _, ok := tunnel.Groups[ga]; !ok {
		tunnel.Groups[ga] = []groupLink{}
	}

	tunnel.Groups[ga] = append(tunnel.Groups[ga], groupLink{
		Name:   n,
		Type:   t,
		Device: d,
	})
	tunnel.Unlock()

	d.State[n] = "-"
}
