package main

import (
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/posener/wstest"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/models/devices"
	"github.com/stretchr/testify/assert"
)

func TestDownloadCA(t *testing.T) {

	main, cleanup := setupServer(t)
	defer cleanup()
	w := makeRequest(t, main.HTTPServer, "GET", "http://localhost/ca.crt", nil)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "application/x-x509-ca-cert", resp.Header.Get("Content-Type"))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, string(body), "BEGIN CERTIFICATE")
}

func TestInsecureWebsocket(t *testing.T) {

	main, cleanup := setupServer(t)
	defer cleanup()

	d := wstest.NewDialer(main.HTTPServer)
	d.Subprotocols = []string{"node"}
	c, _, err := d.Dial("ws://example.org/ws", nil)
	if err != nil {
		t.Fatal(err)
	}

	msgType, msgByte, err := c.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	msg := string(msgByte)

	assert.Equal(t, websocket.TextMessage, msgType)
	assert.Contains(t, msg, `"type":"server-info"`)
	assert.Contains(t, msg, `"uuid":"123"`)
	assert.Contains(t, msg, `"name":"testserver"`)

	// Verify store has saved the connection
	assert.Len(t, main.Store.GetConnections(), 1)
	for _, v := range main.Store.GetConnections() {
		assert.Equal(t, "node", v.Attributes["protocol"])
	}
	c.Close()
	//time.Sleep(200 * time.Millisecond) //TODO better way than sleep?

	waitFor(t, 1*time.Second, "connections should be zero after connection close", func() bool {
		return len(main.Store.GetConnections()) == 0
	})
}

// TestInsecureWebsocketRequestCertificate is a full end to end test between a node and the server going through a node initial connection process etc
func TestInsecureWebsocketRequestCertificate(t *testing.T) {
	main, node, cleanup := setupWebsocketTest(t)
	defer cleanup()

	err := node.Connect()
	assert.NoError(t, err)

	waitFor(t, 1*time.Second, "nodes should be 1", func() bool {
		return len(main.Store.GetNodes()) == 1
	})

	assert.Contains(t, main.Store.GetNodes(), node.UUID)
	assert.Len(t, main.Store.GetConnections(), 1)
	assert.Equal(t, true, main.Store.GetNode(node.UUID).Connected())

	go func() {
		<-time.After(50 * time.Millisecond)
		node.Stop()
	}()
	node.Wait()

	waitFor(t, 1*time.Second, "connections should be 0", func() bool {
		return len(main.Store.GetConnections()) == 0
	})
	assert.Len(t, main.Store.GetConnections(), 0)
	assert.Equal(t, false, main.Store.GetNode(node.UUID).Connected())
}

func TestNodeToServerDevices(t *testing.T) {
	main, node, cleanup := setupWebsocketTest(t)
	defer cleanup()

	err := node.Connect()
	assert.NoError(t, err)

	dev1 := &devices.Device{
		Name: "Device1",
		ID: devices.ID{
			ID: "1",
		},
		Online: true,
		Traits: []string{"OnOff"},
		State: devices.State{
			"on": false,
		},
	}
	node.AddOrUpdate(dev1)
	waitFor(t, 1*time.Second, "should have some devices", func() bool {
		return len(main.Store.Devices.All()) != 0
	})

	//log.Println("devs", main.Store.Devices.All())

	//Make sure node and server has the correct device key which is unique with nodeuuid + device id
	assert.Contains(t, main.Store.Devices.All(), devices.ID{Node: node.UUID, ID: "1"})
	assert.Contains(t, node.Devices.All(), devices.ID{Node: node.UUID, ID: "1"})

}
