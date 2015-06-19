package main

import (
	"time"

	"github.com/goburrow/modbus"
)

type Modbus struct {
	client  modbus.Client
	handler *modbus.RTUClientHandler
}

func (m *Modbus) Connect() error {
	// Modbus RTU/ASCII
	handler := modbus.NewRTUClientHandler("/dev/ttyUSB0")
	handler.BaudRate = 9600
	handler.DataBits = 8
	handler.Parity = "N"
	handler.StopBits = 2
	handler.SlaveId = 1
	handler.Timeout = 5 * time.Second
	//handler.Logger = log.New(os.Stdout, "test: ", log.LstdFlags)
	m.handler = handler
	if err := handler.Connect(); err != nil {
		return err
	}

	m.client = modbus.NewClient(handler)
	return nil
}
func (m *Modbus) Close() error {
	return m.handler.Close()
}
func (m *Modbus) ReadInputRegister(address uint16) ([]byte, error) {
	return m.client.ReadInputRegisters(address-1, 1)
}
