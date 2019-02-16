// Package main provides ...
package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/signal"
	"syscall"

	"github.com/jonaz/goenocean"
	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models/devices"
	"github.com/stampzilla/stampzilla-go/pkg/node"
)

var globalState *State
var enoceanSend chan goenocean.Encoder
var usb300SenderId [4]byte

func main() {
	globalState = NewState()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)

	node := node.New("enocean")
	node.OnConfig(updatedConfig(node))
	wait := node.WaitForFirstConfig()

	node.OnRequestStateChange(onRequestStateChange(node))

	err := node.Connect()
	if err != nil {
		logrus.Error(err)
		return
	}

	logrus.Info("Waiting for config from server")
	err = wait()
	if err != nil {
		logrus.Error(err)
		return
	}

	checkDuplicateSenderIds()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err = setupEnoceanCommunication(ctx, node)
	if err != nil {
		logrus.Error(err)
		return
	}

	node.Wait()
}

func onRequestStateChange(node *node.Node) func(state devices.State, device *devices.Device) error {
	return func(state devices.State, device *devices.Device) error {
		logrus.Debug("OnRequestStateChange:", state, device.ID)

		// Load device config from the config struct
		eDev := globalState.DeviceByString(device.ID.ID)

		if eDev == nil {
			return fmt.Errorf("device %s not found in our config", device.ID.ID)
		}

		// Require a device config for node 4 only
		state.Bool("on", func(on bool) {
			if on {
				logrus.Debugf("turning on %s with senderid %s\n", device.ID.String(), eDev.SenderId)
				eDev.CmdOn()
				return
			}
			logrus.Debugf("turning off %s with senderid %s\n", device.ID.String(), eDev.SenderId)
			eDev.CmdOff()
		})
		state.Float("brightness", func(v float64) {
			bri := int(math.Round(255 * v))
			logrus.Debugf("dimming %s with senderid %s to: %d\n", device.ID.String(), eDev.SenderId, bri)
			eDev.CmdDim(bri)
		})

		return nil
	}

}

func updatedConfig(node *node.Node) func(data json.RawMessage) error {
	return func(data json.RawMessage) error {
		logrus.Debug("Received config from server:", string(data))

		err := json.Unmarshal(data, &globalState.Devices)
		if err != nil {
			return err
		}
		for _, dev := range globalState.Devices {
			node.AddOrUpdate(&devices.Device{
				Type:   getDeviceType(dev),
				Name:   dev.Name,
				ID:     devices.ID{ID: dev.IdString()},
				Online: true,
				Traits: []string{"OnOff", "Brightness"},
				State: devices.State{
					"on": false,
				},
			})
		}
		return nil
	}
}

func getDeviceType(d *Device) string {
	if d.HasSingleRecvEEP("f60201") {
		return "switch"
	}
	return "light"
}

func checkDuplicateSenderIds() {
	for _, d := range globalState.Devices {
		id1 := d.Id()[3] & 0x7f
		for _, d1 := range globalState.Devices {
			if d.Id() == d1.Id() {
				continue
			}
			id2 := d1.Id()[3] & 0x7f
			if id2 == id1 {
				logrus.Error("DUPLICATE ID FOUND when generating senderIds for eltako devices")
			}
		}
	}
}

func setupEnoceanCommunication(ctx context.Context, node *node.Node) error {

	enoceanSend = make(chan goenocean.Encoder, 100)
	recv := make(chan goenocean.Packet, 100)
	err := goenocean.Serial("/dev/ttyUSB0", enoceanSend, recv)
	if err != nil {
		return err
	}

	getIDBase()
	go reciever(ctx, node, recv)
	return nil
}

func getIDBase() {
	p := goenocean.NewPacket()
	p.SetPacketType(goenocean.PacketTypeCommonCommand)
	p.SetData([]byte{0x08})
	enoceanSend <- p
}

func reciever(ctx context.Context, node *node.Node, recv chan goenocean.Packet) {
	for {
		select {
		case p := <-recv:
			if p.PacketType() == goenocean.PacketTypeResponse && len(p.Data()) == 5 {
				copy(usb300SenderId[:], p.Data()[1:4])
				logrus.Debugf("senderid: % x ( % x )", usb300SenderId, p.Data())
				continue
			}
			if p.SenderId() != [4]byte{0, 0, 0, 0} {
				incomingPacket(node, p)
			}
		case <-ctx.Done():
			return
		}
	}
}

func incomingPacket(node *node.Node, p goenocean.Packet) {

	var d *Device
	if d = globalState.Device(p.SenderId()); d == nil {
		//Add unknown device
		d = globalState.AddDevice(p.SenderId(), "UNKNOWN", nil, false)
		logrus.Infof("Found new device %v. Please configure it in the config", d)
	}

	logrus.Debug("Incoming packet")
	if t, ok := p.(goenocean.Telegram); ok {
		logrus.Debug("Packet is goenocean.Telegram")
		for _, deviceEep := range d.RecvEEPs {
			if deviceEep[0:2] != hex.EncodeToString([]byte{t.TelegramType()}) {
				logrus.Trace("Packet is wrong deviceEep ", deviceEep, t.TelegramType())
				continue
			}

			if h := handlers.getHandler(deviceEep); h != nil {
				h.Process(d, t)
				logrus.Debug("Incoming packet processed from", d.IdString())
				//connection.Send(node)
				newState := devices.State{
					"on":         d.On(),
					"brightness": d.Dim,
				}
				node.UpdateState(d.IdString(), newState)
				return
			}
		}
	}

	//fmt.Println("Unknown packet")

}
