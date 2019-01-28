package main

type Config struct {
	Registers Registers
	Device    string
}

func NewConfig() *Config {
	return &Config{
		Registers: make(Registers),
	}

}

type Register struct {
	Name  string
	Id    uint16
	Value interface{}
	Base  int64
}

type Registers map[string]*Register
