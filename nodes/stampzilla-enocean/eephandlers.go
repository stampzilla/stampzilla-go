package main

import (
	"fmt"

	"github.com/jonaz/goenocean"
)

type baseHandler struct {
}

func (handler *baseHandler) On(d *Device) {
	//NOOP
}
func (handler *baseHandler) Off(d *Device) {
	//NOOP
}
func (handler *baseHandler) Toggle(d *Device) {
	//NOOP
}
func (handler *baseHandler) Dim(lvl int, d *Device) {
	//NOOP
}
func (handler *baseHandler) Process(d *Device, t goenocean.Telegram) {
	//NOOP
}

type handlerEepa53808 struct {
	baseHandler
}

func (handler *handlerEepa53808) On(d *Device) {
	p := goenocean.NewEepA53808()
	p.SetDestinationId(d.Id())
	p.SetCommand(2)
	p.SetDimValue(255)
	enoceanSend <- p
}
func (handler *handlerEepa53808) Off(d *Device) {
	p := goenocean.NewEepA53808()
	p.SetDestinationId(d.Id())
	p.SetCommand(2)
	p.SetDimValue(0)
	enoceanSend <- p
}
func (handler *handlerEepa53808) Toggle(d *Device) {
	if d.State == "ON" {
		handler.Off(d)
	} else {
		handler.On(d)
	}
}

type handlerEepa51201 struct {
	baseHandler
}

func (h *handlerEepa51201) Process(d *Device, t goenocean.Telegram) {
	eep := goenocean.NewEepA51201()
	eep.SetTelegram(t) //THIS IS COOL!
	d.SetPower(eep.MeterReading())
	fmt.Println("METERREADING:", eep.MeterReading(), eep.DataType())
	d.PowerUnit = eep.DataType()
	serverSendChannel <- node
}

type handlerEepd20109 struct {
	baseHandler
}

func (h *handlerEepd20109) Process(d *Device, t goenocean.Telegram) {
	eep := goenocean.NewEepD20109()
	eep.SetTelegram(t) //THIS IS COOL!
	fmt.Println("OUTPUTVALUE", eep.OutputValue())
	if eep.CommandId() == 4 {
		if eep.OutputValue() > 0 {
			d.State = "ON"
			//d.State = "ON"
		} else {
			//d.State = "OFF"
			d.State = "OFF"
		}
	}

	//TODO only send update if our values have accually changed
	serverSendChannel <- node
}
