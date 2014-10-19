package main

import (
	"github.com/jonaz/goenocean"
	"github.com/stampzilla/stampzilla-go/protocol"
)

var handlers *eepHandlers

type eepHandlers struct {
	handlers map[string]Handler
}
type Handler interface {
	On(*Device)
	Off(*Device)
	Toggle(*Device)
	Dim(int, *Device)
	Process(*Device, goenocean.Telegram)
	SendElements(*Device) []*protocol.Element
	ReceiveElements(*Device) []*protocol.Element
}

func (h *eepHandlers) getHandler(t string) Handler {
	if handler, ok := h.handlers[t]; ok {
		return handler
	}
	return nil
}

func init() {
	handlers = &eepHandlers{make(map[string]Handler)}
	handlers.handlers["a53808"] = &handlerEepa53808{}
	handlers.handlers["d20109"] = &handlerEepd20109{}
	handlers.handlers["a51201"] = &handlerEepa51201{}
	handlers.handlers["f60201"] = &handlerEepf60201{}
}
