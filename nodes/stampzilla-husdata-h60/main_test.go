package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/e2e"
	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/models/devices"
	"github.com/stretchr/testify/assert"
)

func TestUpdateState(t *testing.T) {
	main, _, cleanup := e2e.SetupWebsocketTest(t)
	defer cleanup()
	e2e.AcceptCertificateRequest(t, main)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Assert we called our heatpump with the correct parameters
		assert.Equal(t, "/api/set?idx=0203&val=225", r.URL.String())
	}))
	defer ts.Close()

	config := NewConfig()
	config.Host = ts.URL
	node := setupNode(config)

	err := node.Connect()
	assert.NoError(t, err)

	dev := &devices.Device{
		Name:   "heatpump",
		Type:   "sensor",
		ID:     devices.ID{ID: "1"},
		Online: true,
		Traits: []string{"TemperatureControl"},
		State:  make(devices.State),
	}
	node.AddOrUpdate(dev)

	b := []byte(fmt.Sprintf(`
		{
		    "type": "state-change",
		    "body": {
		        "%s.1": {
		            "type": "light",
		            "id": "%s.1",
		            "name": "heatpump",
		            "online": true,
		            "state": {
		                "RoomTempSetpoint": 22.5
		            }
		        }
		    }
		}
			`, node.UUID, node.UUID))

	err = node.Client.WriteMessage(websocket.TextMessage, b)
	assert.NoError(t, err)
	e2e.WaitFor(t, 1*time.Second, "wait for node to have updated RoomTempSetpoint", func() bool {
		dev := node.GetDevice("1")
		if dev != nil && dev.State["RoomTempSetpoint"] == 22.5 {
			return true
		}

		return false
	})
}
