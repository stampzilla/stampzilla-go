package websocket

import "errors"

type Router struct {
	callbacks       map[string]func(*Message)
	onClientConnect []func() *Message
}

func NewRouter() *Router {
	r := &Router{}
	r.callbacks = make(map[string]func(*Message))
	return r
}

func (w *Router) AddRoute(route string, handler func(*Message)) {
	w.callbacks[route] = handler
}
func (w *Router) AddClientConnectHandler(handler func() *Message) {
	w.onClientConnect = append(w.onClientConnect, handler)
}

func (w *Router) RunOnClientConnectHandlers() []*Message {
	var msg []*Message
	for _, callback := range w.onClientConnect {
		msg = append(msg, callback())
	}
	return msg
}
func (w *Router) Run(msg *Message) error {
	if cb, ok := w.callbacks[msg.Type]; ok {
		cb(msg)
		return nil
	}
	return errors.New("Undefined websocket route: " + msg.Type)
}
