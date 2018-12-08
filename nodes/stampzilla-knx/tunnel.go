package main

import (
	"fmt"
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
			err := tunnel.DecodeKNX(msg)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"dest": msg.Destination.String(),
				}).Warnf("Failed to handle message: %+v:", msg)
			}

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

func (tunnel *tunnel) DecodeKNX(msg knx.GroupEvent) error {
	tunnel.RLock()
	defer tunnel.RUnlock()

	links, ok := tunnel.Groups[msg.Destination.String()]
	if !ok {
		return fmt.Errorf("No link was found for: %s", msg.Destination.String())
	}

	for _, gl := range links {
		logrus.WithFields(logrus.Fields{
			"dest":     msg.Destination.String(),
			"name":     gl.Name,
			"deviceId": gl.Device.ID,
		}).Trace("Found link")

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
				return fmt.Errorf("Failed to unpack: %s", err.Error())
			}

			gl.Device.State[gl.Name] = dptv
			tunnel.Node.AddOrUpdate(gl.Device)
		} else {
			return fmt.Errorf("Unsupported type: %s", gl.Type)
		}
	}
	return nil
}

func (tunnel *tunnel) AddLink(ga string, n string, t string, d *models.Device) {
	logrus.WithFields(logrus.Fields{
		"dest": ga,
		"name": n,
		"to":   d.ID,
	}).Tracef("Add link")

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

func (tunnel *tunnel) Close() {
	tunnel.Client.Close()
}
