package nibe

import (
	"fmt"
	"sync"
	"time"

	//"github.com/fatih/color".
	"github.com/tarm/serial"
)

type Nibe struct {
	Port string

	onUpdate []func(reg uint16, value int16)

	stop       chan struct{}
	read       chan request
	readResult chan uint16
	write      chan request
	wg         sync.WaitGroup

	sync.RWMutex
}

type request struct {
	register uint16
	value    int16
	result   chan uint16
}

type Message struct {
	Addr byte
	Cmd  byte
	Data []byte
}

var devices = map[byte]string{
	0xC0: "PUMP",
	0x14: "RCU10",
	0x19: "RMU40",
	0x20: "MODBUS40",
	0x22: "SAM40",
}

func New() *Nibe {
	return &Nibe{
		onUpdate: make([]func(reg uint16, value int16), 0),
		read:     make(chan request),
		write:    make(chan request),
	}
}

func (n *Nibe) Connect(p string) {
	n.Stop()

	n.Lock()
	n.Port = p
	n.stop = make(chan struct{})
	n.wg.Add(1)
	n.Unlock()

	go n.Worker()
}

func (n *Nibe) Stop() {
	n.Lock()
	defer n.Unlock()
	if n.stop != nil {
		close(n.stop)
		n.wg.Wait()
		n.stop = nil
	}
}

func (n *Nibe) OnUpdate(cb func(reg uint16, value int16)) {
	n.Lock()
	n.onUpdate = append(n.onUpdate, cb)
	n.Unlock()
}

func (n *Nibe) Read(reg uint16) (uint16, error) {
	result := make(chan uint16)
	n.read <- request{
		register: reg,
		result:   result,
	}

	select {
	case res := <-result:
		return res, nil
	case <-time.After(time.Second * 5):
		return 0, fmt.Errorf("Timeout")
	}
}

func (n *Nibe) Write(reg uint16, value int16) (uint16, error) {
	result := make(chan uint16)
	n.write <- request{
		register: reg,
		value:    value,
		result:   result,
	}

	select {
	case res := <-result:
		return res, nil
	case <-time.After(time.Second * 5):
		return 0, fmt.Errorf("Timeout")
	}
}

func wr(s *serial.Port, data []byte) {
	// sendLog := color.New(color.FgRed, color.Bold)
	// sendLog.Printf(" -> WRITE %x\n", data)
	s.Write(data)
}
