package main

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/models/devices"
	"github.com/stampzilla/stampzilla-go/v2/pkg/node"
	"github.com/vapourismo/knx-go/knx"
	"github.com/vapourismo/knx-go/knx/cemi"
	"github.com/vapourismo/knx-go/knx/dpt"
)

type tunnel struct {
	Node          *node.Node
	Address       string
	Client        *knx.GroupTunnel
	Groups        map[string]groupLink
	OnConnect     func()
	OnDisconnect  func()
	reconnect     bool
	connected     bool
	wg            sync.WaitGroup
	send          chan sendCh
	receive       chan cemi.GroupAddr
	changeAddress chan string
	sync.RWMutex
}

type sendCh struct {
	event knx.GroupEvent
	error chan error
}

type groupLink struct {
	Name   string
	Type   string
	Device *devices.Device
}

func newTunnel(node *node.Node) *tunnel {
	return &tunnel{
		Node:          node,
		Groups:        make(map[string]groupLink),
		reconnect:     true,
		send:          make(chan sendCh),
		receive:       make(chan cemi.GroupAddr),
		changeAddress: make(chan string),
	}
}

func (tunnel *tunnel) Send(event knx.GroupEvent) error {
	errCh := make(chan error)
	s := sendCh{
		event: event,
		error: errCh,
	}
	select {
	case tunnel.send <- s:
	case <-time.After(time.Second * 5):
		return fmt.Errorf("timeout sending after 5 seconds")
	}
	return <-errCh
}

func (tunnel *tunnel) SendWait(event knx.GroupEvent) error {
	errCh := make(chan error)
	s := sendCh{
		event: event,
		error: errCh,
	}
	select {
	case tunnel.send <- s:
	case <-time.After(time.Second * 5):
		return fmt.Errorf("timeout sending after 5 seconds")
	}

	err := <-errCh
	if err != nil {
		return err
	}

	after := time.After(time.Millisecond * 500)
	for {
		select {
		case <-after:
			return fmt.Errorf("timeout waiting for response after 500ms")
		case addr := <-tunnel.receive:
			if addr == event.Destination {
				return nil
			}
		}
	}
}

func (tunnel *tunnel) Connected() bool {
	tunnel.RLock()
	defer tunnel.RUnlock()
	return tunnel.connected
}

func (tunnel *tunnel) Wait() {
	tunnel.wg.Wait()
}

func (tunnel *tunnel) SetAddress(address string) {
	if address == tunnel.Address {
		return
	}
	tunnel.changeAddress <- address
}

// Start handles connection and reconnect on errors or address change.
func (tunnel *tunnel) Start(ctx context.Context) {
	tunnel.wg.Add(1)
	go func() {
		defer tunnel.wg.Done()
		for {
			select {
			case address := <-tunnel.changeAddress:
				tunnel.Address = address
				tunnel.Close()
			case <-ctx.Done():
				return
			default:
			}

			if tunnel.Address == "" {
				logrus.Debug("no tunnel address yet. Sleeping for 2 seconds")
				time.Sleep(time.Second * 2)

				continue
			}

			err := tunnel.connect(ctx, tunnel.Address)
			if err != nil {
				logrus.Error("error connecting to tunnel:", err)
			}

			select {
			case <-ctx.Done():
				return
			default:
			}

			time.Sleep(10 * time.Second)
		}
	}()
}

func (tunnel *tunnel) connect(ctx context.Context, address string) error {
	// Connect to the gateway
	logrus.Infof("Connecting to KNX gateway: %s", address)
	client, err := knx.NewGroupTunnel(address, knx.DefaultTunnelConfig)
	if err != nil {
		return err
	}

	tunnel.Lock()
	tunnel.connected = true
	tunnel.Client = &client
	tunnel.Unlock()

	defer func() {
		tunnel.onDisconnect()
		tunnel.Close()
	}()

	go tunnel.onConnect()
	for {
		select {
		case <-ctx.Done():
			logrus.Info("stopping inbound listener")
			return nil
		case s := <-tunnel.send:
			s.error <- tunnel.Client.Send(s.event)
		case msg, ok := <-client.Inbound():
			if !ok {
				logrus.Error("tunnel disconnected, inbound channel closed")
				return nil
			}
			select {
			case tunnel.receive <- msg.Destination:
			default:
			}
			err := tunnel.decodeKNX(msg)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"dest":    msg.Destination.String(),
					"error":   err,
					"message": fmt.Sprintf("%+v", msg),
				}).Warn("Failed to handle message")
			}
		}
	}
}

func (tunnel *tunnel) onConnect() {
	time.Sleep(time.Second)
	logrus.Info("Connected to KNX gateway")
	// Trigger a read on each group address that we monitor
	tunnel.RLock()
	for ga := range tunnel.Groups {
		tunnel.triggerRead(ga)
	}
	tunnel.RUnlock()
	tunnel.OnConnect()
}

func (tunnel *tunnel) onDisconnect() {
	logrus.Warn("Disconnected from KNX gateway")
	tunnel.Lock()
	tunnel.connected = false
	tunnel.Unlock()
	tunnel.OnDisconnect()
}

func (tunnel *tunnel) triggerRead(ga string) {
	if !tunnel.Connected() { // Dont try to send if we are not connected
		return
	}

	addr, err := cemi.NewGroupAddrString(ga)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"group_address": ga,
			"error":         err,
		}).Error("Failed to read group address")
	}

	err = tunnel.SendWait(knx.GroupEvent{
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
}

func (tunnel *tunnel) decodeKNX(msg knx.GroupEvent) error {
	tunnel.RLock()
	groupAddress, ok := tunnel.Groups[msg.Destination.String()]
	tunnel.RUnlock()
	if !ok {
		return fmt.Errorf("no link was found for: %s", msg.Destination.String())
	}

	logrus.WithFields(logrus.Fields{
		"dest":     msg.Destination.String(),
		"name":     groupAddress.Name,
		"deviceId": groupAddress.Device.ID,
	}).Trace("Found link")

	var value dpt.DatapointValue
	var err error
	switch groupAddress.Type {
	case "bool":
		value = new(dpt.DPT_1001)
	case "procentage":
		value = new(dpt.DPT_5001)
	case "temperature":
		value = new(dpt.DPT_9001) // 2 bytes floating point
	case "lux":
		value = new(dpt.DPT_9004) // 2 bytes floating point
	case "humidity":
		value = new(dpt.DPT_9001) // 2 bytes floating point
	case "co2":
		value = new(dpt.DPT_9001) // 2 bytes floating point
	case "voc":
		value = new(dpt.DPT_9001) // 2 bytes floating point
	case "dewPoint":
		value = new(dpt.DPT_9001) // 2 bytes floating point
	default:
		return fmt.Errorf("unsupported type: %s", groupAddress.Type)
	}

	err = value.Unpack(msg.Data)
	if err != nil {
		return fmt.Errorf("failed to unpack: %s", err.Error())
	}

	newState := make(devices.State)

	// Handle casting of procentage and bool
	switch v := value.(type) {
	case *dpt.DPT_5001:
		newState[groupAddress.Name] = math.Floor(float64(*v)) / 100
	case *dpt.DPT_1001:
		newState[groupAddress.Name] = bool(*v)
	default:
		newState[groupAddress.Name] = v
	}
	groupAddress.Device.SetOnline(true)

	// If temperature and relative humidity is known, calculate absolute humidity
	// Seems unused, will comment it out
	/*
		dh, okh := groupAddress.Device.State["humidity"]
		dt, okt := groupAddress.Device.State["temperature"]
		if okh && okt {
			srh, _ := dh.(dpt.DPT_9001)
			st, _ := dt.(dpt.DPT_9001)

			//pws := 6.116441 * 10 * ((7.591386 - t) / (t + 240.7263))
			//pw := pws * h / 100
			//a := 2.116679 * pw / (273.15 + t)
			//logrus.Warnf("absolute humidity %.2f", a)
			////groupAddress.Device.State["absolute_humidity"] = a

			t := float64(st)
			rh := float64(srh)
			mw := 18.01534                                                           // molar mass of water g/mol
			r := 8.31447215                                                          // Universal gas constant J/mol/K
			ah := (6.112 * math.Exp((17.67*t)/(t+243.5)) * rh * mw) / (273.15 + t*r) // in grams/m^3
			logrus.Warnf("Got temp %s (%#v) and humid %s (%#v) %#v -> %f", st, dt, srh, dh, groupAddress.Device.State, ah)
		}
	*/
	tunnel.Node.UpdateState(groupAddress.Device.ID.ID, newState)
	return nil
}

func (tunnel *tunnel) ClearAllLinks() {
	tunnel.Lock()
	tunnel.Groups = make(map[string]groupLink)
	tunnel.Unlock()
}

func (tunnel *tunnel) AddLink(ga string, n string, t string, d *devices.Device) {
	logrus.WithFields(logrus.Fields{
		"dest": ga,
		"name": n,
		"to":   d.ID,
	}).Tracef("Add link")

	tunnel.Lock()
	tunnel.Groups[ga] = groupLink{
		Name:   n,
		Type:   t,
		Device: d,
	}
	tunnel.Unlock()

	// Trigger a read of the group
	if tunnel.Connected() {
		tunnel.triggerRead(ga)
	}
}

func (tunnel *tunnel) Close() {
	if tunnel.Client != nil {
		tunnel.Client.Close()
	}
}
