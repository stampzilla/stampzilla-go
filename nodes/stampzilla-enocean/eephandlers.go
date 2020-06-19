package main

import (
	"fmt"

	"github.com/jonaz/goenocean"
)

type baseHandler struct {
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
func (h *baseHandler) Learn(d *Device) {
	//NOOP
}
func (h *baseHandler) Dim(lvl int, d *Device) {
	//NOOP
}
func (h *baseHandler) Process(d *Device, t goenocean.Telegram) {
	//NOOP
}
func (h *baseHandler) generateSenderId(d *Device) [4]byte {
	senderId := usb300SenderId
	//id := d.Id()[3]
	//senderId[3] = id & 0x7f

	senderId[3] = byte(d.UniqueId)

	fmt.Printf("Sending with ID:% x\n", senderId)
	return senderId
}

//Handler for profile f60201
type handlerEepf60201 struct {
	baseHandler
}

func (h *handlerEepf60201) On(d *Device) {
	p := goenocean.NewEepF60201()
	p.SetSenderId(h.generateSenderId(d))
	p.SetDestinationId(d.Id())
	//TODO create set methods in EepF60201
	p.SetTelegramData([]byte{0x50}) //ON
	fmt.Printf("Sending ON: % x\n", p.Encode())
	enoceanSend <- p
}
func (h *handlerEepf60201) Off(d *Device) {
	p := goenocean.NewEepF60201()
	p.SetSenderId(h.generateSenderId(d))
	p.SetDestinationId(d.Id())
	//TODO create set methods in EepF60201
	p.SetTelegramData([]byte{0x70}) //OFF
	fmt.Printf("Sending OFF: % x\n", p.Encode())
	enoceanSend <- p
}
func (h *handlerEepf60201) Toggle(d *Device) {
	if d.On() {
		h.Off(d)
	} else {
		h.On(d)
	}
}
func (h *handlerEepf60201) Process(d *Device, t goenocean.Telegram) {
	if t, ok := t.(goenocean.TelegramRps); ok {
		eep := goenocean.NewEepF60201()
		eep.SetTelegram(t) //THIS IS COOL!

		d.Status = ""

		switch {
		case eep.R1A0():
			d.Status = "R1A0"
		case eep.R1A1():
			d.Status = "R1A1"
		case eep.R1B0():
			d.Status = "R1B0"
		case eep.R1B1():
			d.Status = "R1B1"
		case eep.R2B0():
			d.Status = "R2B0"
		case eep.R2B1():
			d.Status = "R2B1"
		}

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
	}
}

//Handler for profile f60201eltako
type handlerEepf60201eltako struct {
	handlerEepf60201
}

func (h *handlerEepf60201eltako) Process(d *Device, t goenocean.Telegram) {
	if t, ok := t.(goenocean.TelegramRps); ok {
		eep := goenocean.NewEepF60201()
		eep.SetTelegram(t)

		//i know this is backwards... eltako is!
		if eep.R1B0() { //ON
			d.SetOn(true)
		}
		if eep.R1B1() { //OFF
			d.SetOn(false)
		}
	}
}

//Handler for profile a53808
type handlerEepa53808 struct {
	baseHandler
}

func (h *handlerEepa53808) On(d *Device) {
	p := goenocean.NewEepA53808()
	p.SetDestinationId(d.Id())
	p.SetCommand(2)
	p.SetDimValue(255)
	p.SetSwitchingCommand(1)
	fmt.Printf("Sending ON: % x\n", p.Encode())
	enoceanSend <- p
}
func (h *handlerEepa53808) Off(d *Device) {
	p := goenocean.NewEepA53808()
	p.SetDestinationId(d.Id())
	p.SetCommand(2)
	p.SetDimValue(0)
	p.SetSwitchingCommand(0)
	fmt.Printf("Sending OFF: % x\n", p.Encode())
	enoceanSend <- p
}
func (h *handlerEepa53808) Toggle(d *Device) {
	if d.On() {
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

	//Simple learn. Set learn bit to false and send
	p1 := goenocean.NewEepA53808()
	tmp := p1.TelegramData()
	tmp[3] &^= 0x08
	tmp[3] |= (0 << 3) & 0x08
	p1.SetTelegramData(tmp)
	p1.SetCommand(2)
	fmt.Printf("Sending learn simple: % x\n", p1.Encode())
	enoceanSend <- p1
}

//Handler for profile a53808eltako
type handlerEepa53808eltako struct {
	handlerEepa53808
}

func (h *handlerEepa53808eltako) Toggle(d *Device) {
	if d.On() {
		h.Off(d)
	} else {
		h.On(d)
	}
}
func (h *handlerEepa53808eltako) On(d *Device) {
	p := goenocean.NewEepA53808()
	p.SetSenderId(h.generateSenderId(d))
	p.SetDestinationId(d.Id())
	p.SetCommand(2)
	p.SetDimValue(255)
	p.SetSwitchingCommand(1)
	fmt.Printf("Sending ON: % x\n", p.Encode())
	enoceanSend <- p
}

func (h *handlerEepa53808eltako) Off(d *Device) {
	p := goenocean.NewEepA53808()
	p.SetSenderId(h.generateSenderId(d))
	p.SetDestinationId(d.Id())
	p.SetCommand(2)
	p.SetDimValue(0)
	p.SetSwitchingCommand(0)
	fmt.Printf("Sending OFF: % x\n", p.Encode())
	enoceanSend <- p
}
func (h *handlerEepa53808eltako) Dim(lvl int, d *Device) {
	p := goenocean.NewEepA53808()
	p.SetSenderId(h.generateSenderId(d))
	p.SetDestinationId(d.Id())
	p.SetCommand(2)
	//TODO dim level should be 0-100 in stampzilla and send 0-255 to enocean device
	p.SetDimValue(uint8(lvl))
	p.SetSwitchingCommand(1)
	fmt.Printf("Sending DIM: % x\n", p.Encode())
	enoceanSend <- p
}
func (h *handlerEepa53808eltako) Process(d *Device, t goenocean.Telegram) {
	eep := goenocean.NewEepA53808()
	eep.SetTelegram(t) //THIS IS COOL!
	fmt.Println("DIMVALUE:", eep.DimValue())
	fmt.Println("SW command STATUS:", eep.SwitchingCommand())
	if eep.SwitchingCommand() == 1 {
		d.SetOn(true)
	}
	if eep.SwitchingCommand() == 0 {
		d.SetOn(false)
	}
	d.Dim = int64(eep.DimValue())
}
func (h *handlerEepa53808eltako) Learn(d *Device) {
	p := goenocean.NewTelegram4bsLearn()
	p.SetLearnFunc(0x38)
	p.SetLearnType(0x08)
	fmt.Printf("Sending learn: % x\n", p.Encode())
	enoceanSend <- p

	//Simple learn. Set learn bit to false and send
	p1 := goenocean.NewEepA53808()
	p.SetSenderId(h.generateSenderId(d))
	tmp := p1.TelegramData()
	tmp[3] &^= 0x08
	tmp[3] |= (0 << 3) & 0x08
	p1.SetTelegramData(tmp)
	p1.SetCommand(2)
	fmt.Printf("Sending learn simple: % x\n", p1.Encode())
	enoceanSend <- p1
}

//Handler for profile a51201
type handlerEepa51201 struct {
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
}

//Handler for profile d20109
type handlerEepd20109 struct {
	baseHandler
}

func (h *handlerEepd20109) Process(d *Device, t goenocean.Telegram) {
	eep := goenocean.NewEepD20109()
	eep.SetTelegram(t) //THIS IS COOL!
	fmt.Println("OUTPUTVALUE", eep.OutputValue())
	if eep.CommandId() == 4 {
		value := eep.OutputValue()
		d.Dim = int64(value)
		if value > 0 {
			d.SetOn(true)
		} else {
			d.SetOn(false)
		}
	}
}
