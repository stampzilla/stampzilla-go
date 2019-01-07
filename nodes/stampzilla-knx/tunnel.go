package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/models/devices"
	"github.com/stampzilla/stampzilla-go/pkg/node"
	"github.com/vapourismo/knx-go/knx"
	"github.com/vapourismo/knx-go/knx/cemi"
	"github.com/vapourismo/knx-go/knx/dpt"
)

type tunnel struct {
	Node      *node.Node
	Address   string
	Connected bool
	Client    *knx.GroupTunnel

	Groups map[string][]groupLink

	OnConnect    func()
	OnDisconnect func()

	reconnect bool
	wg        sync.WaitGroup
	sync.RWMutex
}

type groupLink struct {
	Name   string
	Type   string
	Device *devices.Device
}

func newTunnel(node *node.Node) *tunnel {
	return &tunnel{
		Node:      node,
		Groups:    make(map[string][]groupLink),
		reconnect: true,
	}
}

func (tunnel *tunnel) SetAddress(address string) {
	if address == tunnel.Address {
		return
	}

	for {
		err := tunnel.Connect(address)

		if err == nil {
			return
		}

		logrus.WithFields(logrus.Fields{
			"address": address,
			"error":   err,
		}).Error("Failed to open KNX tunnel")

		if !tunnel.reconnect {
			return
		}
		logrus.Warn("Going to try again in 10s")
		<-time.After(time.Second * 10)
	}
}

func (tunnel *tunnel) Connect(address string) error {
	// Connect to the gateway
	client, err := knx.NewGroupTunnel(address, knx.DefaultTunnelConfig)
	if err != nil {
		return err
	}

	// Disconnect the previous tunnel
	if tunnel.Client != nil {
		tunnel.reconnect = false
		tunnel.Client.Close()
		tunnel.wg.Wait()
		tunnel.reconnect = true
	}

	// Start using the new one
	go func() {
		tunnel.wg.Add(1)
		tunnel.Address = address
		tunnel.Client = &client
		tunnel.onConnect()

		defer func() {
			tunnel.onDisconnect()
			tunnel.Client = nil
			client.Close()
			tunnel.wg.Done()
			logrus.Warn("Disconnect from KNX gateway done")

			if tunnel.reconnect {
				logrus.Warn("Going to try again in 10s")
				<-time.After(time.Second * 10)
				go tunnel.Connect(address)
			}
		}()

		for msg := range client.Inbound() {
			err := tunnel.decodeKNX(msg)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"dest":    msg.Destination.String(),
					"error":   err,
					"message": fmt.Sprintf("%+v", msg),
				}).Warn("Failed to handle message")
			}
		}
	}()

	return nil
}

func (tunnel *tunnel) onConnect() {
	logrus.Info("Connected to KNX gateway")
	tunnel.Connected = true
	// Trigger a read on each group address that we monitor
	for ga, _ := range tunnel.Groups {
		tunnel.triggerRead(ga)
	}
	tunnel.OnConnect()
}
func (tunnel *tunnel) onDisconnect() {
	logrus.Warn("Disconnected from KNX gateway")
	tunnel.Connected = false
	tunnel.OnDisconnect()
}

func (tunnel *tunnel) triggerRead(ga string) {
	if !tunnel.Connected { // Dont try to send if we are not connected
		return
	}

	addr, err := cemi.NewGroupAddrString(ga)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"group_address": ga,
			"error":         err,
		}).Error("Failed to read group address")
	}

	err = tunnel.Client.Send(knx.GroupEvent{
		Command:     knx.GroupRead,
		Destination: addr,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"group_address": ga,
			"error":         err,
		}).Error("Failed to send read request")
	}
	logrus.WithFields(logrus.Fields{
		"group_address": ga,
	}).Info("Sent read request")

	<-time.After(time.Millisecond * 100)
}

func (tunnel *tunnel) decodeKNX(msg knx.GroupEvent) error {
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
			gl.Device.Online = true
			tunnel.Node.AddOrUpdate(gl.Device)
		} else {
			return fmt.Errorf("Unsupported type: %s", gl.Type)
		}
	}
	return nil
}

func (tunnel *tunnel) ClearAllLinks() {
	tunnel.Groups = make(map[string][]groupLink)
}

func (tunnel *tunnel) AddLink(ga string, n string, t string, d *devices.Device) {
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

	d.State[n] = nil

	// Trigger a read of the group
	if tunnel.Connected {
		tunnel.triggerRead(ga)
	}
}

func (tunnel *tunnel) Close() {
	if tunnel.Client != nil {
		tunnel.Client.Close()
		tunnel.wg.Wait()
	}
}
