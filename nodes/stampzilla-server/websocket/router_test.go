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

func TestClientConnectHandler(t *testing.T) {

	r := NewRouter()

	count := 0
	handler := func() *Message {
		count++
		return &Message{Type: "test"}
	}

	r.AddClientConnectHandler(handler)
	r.AddClientConnectHandler(handler)

	r.RunOnClientConnectHandlers()

	if count != 2 {
		t.Errorf("Expected 2 handlers to have run. got: %d", count)
	}
}

func TestRunUndefinedRoute(t *testing.T) {
	msg := &Message{Type: "cmd"}
	r := NewRouter()
	err := r.Run(msg)
	t.Log(err)
	if _, ok := err.(*ErrNoSuchRoute); ok {
		return
	}
	t.Error("Expected ErrNoSuchRoute error. got:", err)
}
