package main

import "github.com/jonaz/goenocean"

var handlers *eepHandlers

type eepHandlers struct {
	handlers map[string]func(*Device, goenocean.Telegram)
}

func (h *eepHandlers) getHandler(t string) func(*Device, goenocean.Telegram) {
	if handler, ok := h.handlers[t]; ok {
		return handler
	}
	return nil
}

func setupEepHandlers() {
	handlers = &eepHandlers{make(map[string]func(*Device, goenocean.Telegram))}
	handlers.handlers["d20109"] = eepd20109
	handlers.handlers["a51201"] = eepa51201
}
