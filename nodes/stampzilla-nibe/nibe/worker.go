package nibe

import (
	"bytes"
	"encoding/binary"
	"io"
	"time"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	"github.com/tarm/serial"
)

func (n *Nibe) Worker() {
	for {
		select {
		case <-n.stop:
			n.wg.Done()
			return
		default:
			reset = true
			err := n.connect()
			if err != nil {
				logrus.Warn(err)
				<-time.After(time.Second)
			}
		}
	}
}

func (n *Nibe) connect() error {
	c := &serial.Config{Name: n.Port, Baud: 9600, ReadTimeout: time.Millisecond * 1}
	s, err := serial.OpenPort(c)
	if err != nil {
		return err
	}

	buf := make([]byte, 128)
	accumulated := []byte{}
	for {
		// Should we exit?
		select {
		case <-n.stop:
			return nil
		default:
		}

		length, err := s.Read(buf)
		if err == io.EOF {
			// SEND NEXT MESSAGE
			continue
		}
		if err != nil {
			return err
		}

		accumulated = append(accumulated, buf[:length]...)

		message := &Message{}
		for message != nil {
			message, accumulated = decodeMessage(accumulated)

			if message != nil && message.Addr == 0x20 { // A message to modbus40 (ous)
				n.handle(s, message)
			}
		}
	}
}

var reset = true

func (n *Nibe) handle(s *serial.Port, m *Message) error {
	blue := color.New(color.FgBlue)
	// log := color.New(color.FgWhite)

	switch m.Cmd {
	case 0x68: // Cyclic LOG data
		wr(s, []byte{0x06})

		var parameter struct {
			Id    uint16
			Value int16
		}
		r := bytes.NewReader(m.Data)
		if len(m.Data) == 80 {
			n.RLock()
			for {
				err := binary.Read(r, binary.LittleEndian, &parameter)
				if err != nil {
					break
				}

				for _, cb := range n.onUpdate {
					// Notify that a register value was received
					cb(parameter.Id, parameter.Value)
				}

				//if p, ok := parameters[int(parameter.Id)]; ok {
				//	blue.Printf(" - %d - %f (%s)\n", int(parameter.Id), float64(parameter.Value)/float64(p.Factor), p.Title)
				//} else {
				//log.Printf(" - %d - %d\n", int(parameter.Id), int(parameter.Value))
				//}
			}
			n.RUnlock()
		}
	case 0x69: // READ token
		// log.Printf(" -> READ TOKEN -> %x\n", m.Cmd);
		select {
		case req := <-n.read:
			// Read the requested register
			n.readResult = req.result
			msg := encodeMessage(
				0x69,         // CMD
				req.register, // ADDR
				[]byte{},
			)
			wr(s, msg)
		default:
			// Nothing to read, just ack
			wr(s, []byte{0x06})
		}

	case 0x6A: // Read result
		wr(s, []byte{0x06})

		// blue.Printf(" -> Read result -> %#v\n", m.Data)
		if n.readResult != nil {
			select {
			case n.readResult <- binary.LittleEndian.Uint16(m.Data[2:]):
				// Delivered the result
			default:
				// No one listened
				n.readResult = nil
			}
		}
	case 0x6B: // WRITE token
		if reset {
			data := make([]byte, 4)
			binary.LittleEndian.PutUint32(data[0:], 1)
			msg := encodeMessage(
				0x6B,          // CMD
				uint16(45171), // ADDR
				data,
			)
			wr(s, msg)
			reset = false
		} else {
			wr(s, []byte{0x06})
		}
	case 0x6C: // Write result
		// blue.Printf(" -> Write result -> %#v\n", m.Data)
		wr(s, []byte{0x06})
	case 0x6D: // Connected controller info
		// log.Printf(" -> Controller -> %s\n", string(m.Data[3:]))
		wr(s, []byte{0x06})
	case 0xEE: // ID?

		log5 := color.New(color.FgMagenta)
		log5.Printf(" -> ID? -> %#v\n", m)

		msg := encodeMessage(
			0xEE,      // CMD
			uint16(2), // ADDR
			[]byte{0x4},
		)
		wr(s, msg)
		// wr(s, []byte{0x06});
	default:
		blue.Printf(" -> UNKNOWN -> %#v\n", m)
		wr(s, []byte{0x06})
	}

	return nil
}
