package main

import "sync"

type Config struct {
	Registers Registers
	Device    string
	sync.Mutex
}

// GetRegisters returns a copy of the configured registers safe to use while config is updated.
func (c *Config) GetRegisters() []Register {
	c.Lock()
	defer c.Unlock()
	registers := make([]Register, 0, len(c.Registers))
	for _, r := range c.Registers {
		registers = append(registers, *r)
	}
	return registers
}

func NewConfig() *Config {
	return &Config{
		Registers: make(Registers),
	}
}

type Register struct {
	Name string
	Id   uint16
	Base int64
	// Write marks the register as a writable holding register. It is then read
	// with ReadHoldingRegisters and can be changed through state changes.
	Write bool
}

type Registers map[string]*Register
