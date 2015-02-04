package websocket

import "testing"

func TestRouterAddRoute(t *testing.T) {

	handler := func(msg *Message) {
	}

	r := NewRouter()
	r.AddRoute("cmd", handler)

	msg := &Message{Type: "cmd"}
	if r.Run(msg) == nil {
		return
	}
	t.Error("Route handle for cmd not found")
}
