package main

import (
	"bytes"
	"encoding/json"
	"os"
)

type Register struct {
	Name  string
	Id    uint16
	Value interface{}
	Base  int64
}

type Registers struct {
	Registers map[string]*Register
}

func NewRegisters() *Registers {
	return &Registers{
		Registers: make(map[string]*Register),
	}
}

func (c *Registers) SaveToFile(filepath string) error {
	configFile, err := os.Create(filepath)
	if err != nil {
		return err
	}

	var out bytes.Buffer
	b, err := json.MarshalIndent(c, "", "\t")
	if err != nil {
		return err
	}
	json.Indent(&out, b, "", "\t")
	out.WriteTo(configFile)
	return nil
}
func (c *Registers) ReadFromFile(filepath string) error {
	configFile, err := os.Open(filepath)
	if err != nil {
		return err
	}

	registers := &Registers{}
	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&registers); err != nil {
		return err
	}

	*c = *registers
	return nil
}

func (c *Registers) GetState() interface{} {
	return c
}
