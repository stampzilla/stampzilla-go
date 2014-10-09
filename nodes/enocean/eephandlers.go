package main

import (
	"fmt"

	"github.com/jonaz/goenocean"
)

func eepa51201(d *Device, t goenocean.Telegram) {
	eep := goenocean.NewEepA51201()
	eep.SetTelegram(t) //THIS IS COOL!
	d.SetPower(eep.MeterReading())
	fmt.Println("METERREADING:", eep.MeterReading(), eep.DataType())
	d.PowerUnit = eep.DataType()
	serverSendChannel <- node
}

func eepd20109(d *Device, t goenocean.Telegram) {
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
