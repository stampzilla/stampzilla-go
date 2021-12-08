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
		if r.URL.Query().Get("idx") == "0203" {
			assert.Equal(t, "/api/set?idx=0203&val=225", r.URL.String())
			return
		}
		if r.URL.Query().Get("idx") == "6209" {
			assert.Equal(t, "/api/set?idx=6209&val=2", r.URL.String())
			return
		}
		t.Error("unexpected get parameters")
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
		                "RoomTempSetpoint": 22.5,
		                "ExtraWarmWater": 2
		            }
		        }
		    }
		}
			`, node.UUID, node.UUID))

	err = node.Client.WriteMessage(websocket.TextMessage, b)
	assert.NoError(t, err)
	var syncedDev *devices.Device
	e2e.WaitFor(t, 1*time.Second, "wait for node to have updated RoomTempSetpoint", func() bool {
		syncedDev = node.GetDevice("1")
		if dev != nil && syncedDev.State["RoomTempSetpoint"] != nil {
			return true
		}
		return false
	})
	assert.Equal(t, 22.5, syncedDev.State["RoomTempSetpoint"])
	assert.Equal(t, float64(2), syncedDev.State["ExtraWarmWater"])
}
