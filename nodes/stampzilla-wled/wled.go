package main

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/models/devices"
)

// wledMessage is the shape of JSON pushed by WLED over the WebSocket
// (equivalent to GET /json/si — state + info objects).
type wledMessage struct {
	State struct {
		On  bool `json:"on"`
		Bri int  `json:"bri"`
	} `json:"state"`
}

// briFromFloat converts a 0.0–1.0 brightness to the WLED 0–255 bri integer.
func briFromFloat(b float64) int {
	if b < 0 {
		b = 0
	}
	if b > 1 {
		b = 1
	}
	return int(math.Round(b * 255))
}

// wsURL builds the ws:// (or wss://) URL for the WLED /ws endpoint.
// host may be a bare IP/hostname ("192.168.1.100"), a bare HTTP URL
// ("http://192.168.1.100"), or an already-formed ws:// URL.
func wsURL(host string) string {
	host = strings.TrimRight(host, "/")

	var scheme, rest string
	switch {
	case strings.HasPrefix(host, "https://"):
		scheme, rest = "wss", host[len("https://"):]
	case strings.HasPrefix(host, "http://"):
		scheme, rest = "ws", host[len("http://"):]
	case strings.HasPrefix(host, "wss://"):
		scheme, rest = "wss", host[len("wss://"):]
	case strings.HasPrefix(host, "ws://"):
		scheme, rest = "ws", host[len("ws://"):]
	default:
		scheme, rest = "ws", host
	}

	// Strip any existing path — we always append /ws.
	if i := strings.Index(rest, "/"); i != -1 {
		rest = rest[:i]
	}

	return scheme + "://" + rest + "/ws"
}

// parseWLEDState decodes a raw WLED WebSocket message into a devices.State
// containing "on" (bool) and "brightness" (float64 0.0–1.0).
func parseWLEDState(msg []byte) (devices.State, error) {
	var m wledMessage
	if err := json.Unmarshal(msg, &m); err != nil {
		return nil, fmt.Errorf("wled: parseWLEDState: %w", err)
	}
	return devices.State{
		"on":         m.State.On,
		"brightness": float64(m.State.Bri) / 255.0,
	}, nil
}
