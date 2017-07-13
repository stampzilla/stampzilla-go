package main

import (
	"fmt"
	"log"

	"github.com/gozwave/gozw/application"
	"github.com/gozwave/gozw/frame"
	serialapi "github.com/gozwave/gozw/serial-api"
	"github.com/gozwave/gozw/session"
	"github.com/gozwave/gozw/transport"
	"github.com/gozwave/gozw/util"
	"go.bug.st/serial.v1/enumerator"
)

func connectToController() *application.Layer {

	port := findPort()

	if port == "" {
		err := fmt.Errorf("No known device found, please provide the port address")
		panic(err)
	}

	transport, err := transport.NewSerialPortTransport(port, 115200)
	if err != nil {
		panic(err)
	}

	frameLayer := frame.NewFrameLayer(transport)
	sessionLayer := session.NewSessionLayer(frameLayer)
	apiLayer := serialapi.NewLayer(sessionLayer)
	appLayer, err := application.NewLayer(apiLayer)
	if err != nil {
		panic(err)
	}

	return appLayer
}

func findPort() string {
	knownDevices := util.KnownUsbDevices()

	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		log.Fatal(err)
	}

	for _, port := range ports {
		id := fmt.Sprintf("%s:%s", port.VID, port.PID)

		if dev, ok := knownDevices[id]; ok {
			fmt.Printf("Found port: %s (%s)\n", port.Name, dev)
			return port.Name
		}
	}

	return ""
}
