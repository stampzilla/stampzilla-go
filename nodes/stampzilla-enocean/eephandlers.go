package main

import (
	"fmt"

	"github.com/jonaz/goenocean"
	"github.com/stampzilla/stampzilla-go/protocol"
)

type baseHandler struct { // {{{
}

func (h *baseHandler) On(d *Device) {
	//NOOP
}
func (h *baseHandler) Off(d *Device) {
	//NOOP
}
func (h *baseHandler) Toggle(d *Device) {
	//NOOP
}
func (h *baseHandler) Dim(lvl int, d *Device) {
	//NOOP
}
func (h *baseHandler) Process(d *Device, t goenocean.Telegram) {
	//NOOP
}
func (h *baseHandler) SendElements(d *Device) []*protocol.Element {
	return nil
}
func (h *baseHandler) ReceiveElements(d *Device) []*protocol.Element {
	return nil
}

// }}}

//Handler for profile f60201
type handlerEepf60201 struct { // {{{
	baseHandler
}

func (h *handlerEepf60201) On(d *Device) {
	p := goenocean.NewEepF60201()
	p.SetDestinationId(d.Id())
	//TODO create set methods in EepF60201
	p.SetTelegramData([]byte{0x50}) //ON
	enoceanSend <- p
}
func (h *handlerEepf60201) Off(d *Device) {
	p := goenocean.NewEepF60201()
	p.SetDestinationId(d.Id())
	//TODO create set methods in EepF60201
	p.SetTelegramData([]byte{0x70}) //OFF
	enoceanSend <- p
}
func (h *handlerEepf60201) Toggle(d *Device) {
	if d.On {
		h.Off(d)
	} else {
		h.On(d)
	}
}
func (h *handlerEepf60201) Process(d *Device, t goenocean.Telegram) {
	if t, ok := t.(goenocean.TelegramRps); ok {
		eep := goenocean.NewEepF60201()
		eep.SetTelegram(t) //THIS IS COOL!
		fmt.Println("R1A0", eep.R1A0())
		fmt.Println("R1A1", eep.R1A1())
		fmt.Println("R1B0", eep.R1B0())
		fmt.Println("R1B1", eep.R1B1())
		fmt.Println("R2B0", eep.R2B0())
		fmt.Println("R2B1", eep.R2B1())
		//if eep.CommandId() == 4 {
		//if eep.OutputValue() > 0 {
		//d.State = "ON"
		////d.State = "ON"
		//} else {
		////d.State = "OFF"
		//d.State = "OFF"
		//}
		//}

		//TODO only send update if our values have accually changed
		//serverSendChannel <- node
	}
}

// }}}

//Handler for profile a53808
type handlerEepa53808 struct { // {{{
	baseHandler
}

func (h *handlerEepa53808) On(d *Device) {
	p := goenocean.NewEepA53808()
	p.SetDestinationId(d.Id())
	p.SetCommand(2)
	p.SetDimValue(255)
	enoceanSend <- p
}
func (h *handlerEepa53808) Off(d *Device) {
	p := goenocean.NewEepA53808()
	p.SetDestinationId(d.Id())
	p.SetCommand(2)
	p.SetDimValue(0)
	enoceanSend <- p
}
func (h *handlerEepa53808) Toggle(d *Device) {
	if d.On {
		h.Off(d)
	} else {
		h.On(d)
	}
}
func (h *handlerEepa53808) Learn(d *Device) {
	p := goenocean.NewTelegram4bsLearn()
	p.SetLearnFunc(0x38)
	p.SetLearnType(0x08)

	// OMG THIS WORKS :D:D
	fmt.Printf("Sending learn: % x\n", p.Encode())
	enoceanSend <- p
}

func (h *handlerEepa53808) SendElements(d *Device) []*protocol.Element {
	var els []*protocol.Element

	// TOGGLE
	el := &protocol.Element{}
	el.Type = protocol.ElementTypeToggle
	el.Name = d.Name
	el.Command = &protocol.Command{
		Cmd:  "toggle",
		Args: []string{d.IdString()},
	}
	el.Feedback = `Devices["` + d.IdString() + `"].On`
	els = append(els, el)

	//TODO also generate Off and On here

	return els
}

// }}}

//Handler for profile a51201
type handlerEepa51201 struct { // {{{
	baseHandler
}

func (h *handlerEepa51201) Process(d *Device, t goenocean.Telegram) {
	eep := goenocean.NewEepA51201()
	eep.SetTelegram(t) //THIS IS COOL!
	fmt.Println("METERREADING:", eep.MeterReading(), eep.DataType())
	if eep.DataType() == "W" {
		d.SetPowerW(eep.MeterReading())
	} else {
		d.SetPowerkWh(eep.MeterReading())
	}
	serverSendChannel <- node
}

// }}}

//Handler for profile d20109
type handlerEepd20109 struct { // {{{
	baseHandler
}

func (h *handlerEepd20109) Process(d *Device, t goenocean.Telegram) {
	eep := goenocean.NewEepD20109()
	eep.SetTelegram(t) //THIS IS COOL!
	fmt.Println("OUTPUTVALUE", eep.OutputValue())
	if eep.CommandId() == 4 {
		if eep.OutputValue() > 0 {
			d.On = true
		} else {
			d.On = false
		}
	}

	//TODO only send update if our values have accually changed
	serverSendChannel <- node
}

// }}}
