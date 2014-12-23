package websocket

import (
	"encoding/json"
	"errors"
)

type Router struct {
	callbacks       map[string]func(string)
	onClientConnect []func() *Message
}

func NewRouter() *Router {
	r := &Router{}
	r.callbacks = make(map[string]func(string))
	return r
}

func (w *Router) AddRoute(route string, handler func(string)) {
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
func (w *Router) Run(str string) error {

	var msg *Message
	err := json.Unmarshal([]byte(str), &msg)
	if err != nil {
		return err
	}

	if cb, ok := w.callbacks[msg.Type]; ok {
		cb(str)
		return nil
	}
	return errors.New("Undefined websocket route: " + msg.Type)
}
