package websocket

import "errors"

type Router struct {
	callbacks map[string]func(*Message)
}

func NewRouter() *Router {
	r := &Router{}
	r.callbacks = make(map[string]func(*Message))
	return r
}

func (w Router) AddRoute(route string, handler func(*Message)) {
	w.callbacks[route] = handler
}

func (w Router) Run(msg *Message) error {
	if cb, ok := w.callbacks[msg.Type]; ok {
		cb(msg)
		return nil
	}
	return errors.New("Undefined websocket route: " + msg.Type)
}
