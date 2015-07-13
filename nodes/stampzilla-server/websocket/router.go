package websocket

import "fmt"

type ErrNoSuchRoute struct {
	route string
}

func (self *ErrNoSuchRoute) Error() string {
	return fmt.Sprintf("Undefined websocket route: %s", self.route)
}

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

	//var msg *Message
	//err := json.Unmarshal([]byte(str), &msg)
	//if err != nil {
	//return err
	//}

	if cb, ok := w.callbacks[msg.Type]; ok {
		cb(msg)
		return nil
	}
	return &ErrNoSuchRoute{route: msg.Type}
	//return errors.New("Undefined websocket route: " + msg.Type)
}
