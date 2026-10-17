package node

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/models/devices"
	"github.com/stretchr/testify/assert"
)

type fakeWebsocket struct{}

func (f *fakeWebsocket) OnConnect(cb func()) {}
func (f *fakeWebsocket) ConnectContext(ctx context.Context, addr string, h http.Header) error {
	return nil
}
func (f *fakeWebsocket) ConnectWithRetry(ctx context.Context, addr string, h http.Header) {}
func (f *fakeWebsocket) Wait()                                                            {}
func (f *fakeWebsocket) Read() <-chan []byte                                              { return nil }
func (f *fakeWebsocket) WriteJSON(v interface{}) error                                    { return nil }
func (f *fakeWebsocket) WriteMessage(messageType int, data []byte) error                  { return nil }
func (f *fakeWebsocket) SetTLSConfig(c *tls.Config)                                       {}

func TestOnRequestStateChangeKeepsUnrelatedState(t *testing.T) {
	n := NewWithClient(&fakeWebsocket{})
	dev := &devices.Device{
		ID: devices.ID{ID: "1"},
		State: devices.State{
			"on":          false,
			"temperature": 21.5,
		},
	}
	n.Devices.Add(dev) // AddOrUpdate needs the sync worker running

	var got devices.State
	n.OnRequestStateChange(func(state devices.State, device *devices.Device) error {
		got = state
		return nil
	})

	// a partial request like the ones sent from rules and scenes
	data, err := json.Marshal(map[devices.ID]devices.State{
		dev.ID: {"on": true},
	})
	assert.NoError(t, err)

	for _, cb := range n.getCallbacks()["state-change"] {
		assert.NoError(t, cb(data))
	}

	assert.Equal(t, devices.State{"on": true}, got)
	assert.Equal(t, devices.State{"on": true, "temperature": 21.5}, n.GetDevice("1").State)
}
