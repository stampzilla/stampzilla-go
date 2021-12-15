package main

import (
	"net"
	"testing"
	"time"

	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/e2e"
	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/models/devices"
	"github.com/stretchr/testify/assert"
)

func TestUpdateStateFromUDP(t *testing.T) {
	//logrus.SetLevel(logrus.DebugLevel)
	main, _, cleanup := e2e.SetupWebsocketTest(t)
	defer cleanup()
	e2e.AcceptCertificateRequest(t, main)

	config.EDPPort = "9999"
	_, node, listenPort := start()
	listenPort <- config.EDPPort

	time.Sleep(time.Millisecond * 100) // Wait for udp server to start
	err := writeUDP(config.EDPPort)
	assert.NoError(t, err)

	e2e.WaitFor(t, 1*time.Second, "we should have 1 device", func() bool {
		return len(main.Store.GetDevices().All()) == 1
	})
	//spew.Dump(main.Store.Devices.All())
	//spew.Dump(node.Devices.All())

	// Assert that the device exists in the server after we got UDP packet
	assert.Equal(t, "Zone Kök IR", main.Store.GetDevices().Get(devices.ID{ID: "zone.8", Node: node.UUID}).Name)
	assert.Equal(t, "Zone Kök IR", main.Store.GetDevices().Get(devices.ID{ID: "zone.8", Node: node.UUID}).Name)
}

func writeUDP(port string) error {
	d := []byte{0x45, 0x2, 0x0, 0x3e, 0x0, 0x0, 0x0, 0xe8, 0x3, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x2, 0x0, 0x95, 0xa1, 0x33, 0x0, 0x45, 0x32, 0x5b, 0x23, 0x31, 0x30, 0x30, 0x30, 0x7c, 0x32, 0x31, 0x31, 0x35, 0x35, 0x37, 0x30, 0x33, 0x31, 0x31, 0x32, 0x30, 0x32, 0x30, 0x7c, 0x5a, 0x4f, 0x7c, 0x38, 0x7c, 0x4b, 0xf6, 0x6b, 0x20, 0x49, 0x52, 0xa6, 0x5a, 0x4f, 0x4e, 0x45, 0xa6, 0x31, 0xa6, 0x4c, 0x61, 0x72, 0x6d, 0x7c, 0x7c, 0x30, 0x5d}
	conn, err := net.Dial("udp", "127.0.0.1:"+port)
	if err != nil {
		return err
	}
	_, err = conn.Write(d)
	return err
}
