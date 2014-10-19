package main

import "github.com/stampzilla/stampzilla-go/protocol"

type ElementGenerator struct {
	State *State
	Node  *protocol.Node
}

func (e *ElementGenerator) Run() {
	for _, el := range e.generate() {
		e.Node.AddElement(el)
	}
}
func (e *ElementGenerator) generate() []*protocol.Element {
	var elements []*protocol.Element
	for _, device := range e.State.Devices {
		elements = append(elements, e.device(device)...)
	}
	return elements
}

func (e *ElementGenerator) device(d *Device) []*protocol.Element {

	var elements []*protocol.Element
	for _, name := range d.SendEEPs {
		handler := handlers.getHandler(name)
		elements = append(elements, handler.SendElements(d)...)
	}
	for _, name := range d.RecvEEPs {
		handler := handlers.getHandler(name)
		elements = append(elements, handler.ReceiveElements(d)...)
	}

	return elements
}
