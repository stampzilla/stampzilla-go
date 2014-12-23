package websocket

import (
	"encoding/json"
	"testing"
)

func TestRouterAddRoute(t *testing.T) {

	handler := func(msg string) {
	}

	r := NewRouter()
	r.AddRoute("cmd", handler)

	str, err := json.Marshal(&Message{Type: "cmd"})
	if err != nil {
		t.Error(err)
	}
	if r.Run(string(str)) == nil {
		return
	}
	t.Error("Route handle for cmd not found")
}
