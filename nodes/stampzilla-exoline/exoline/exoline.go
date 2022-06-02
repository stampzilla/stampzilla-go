package exoline

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"

	"github.com/sirupsen/logrus"
)

func FloatTobytes(float float64) []byte {
	bits := math.Float32bits(float32(float))
	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, bits)
	return bytes
}

func AsRoundedFloat(data []byte) (float64, error) {
	// this works OK according to protocol example:
	// Extract temperature 21,8
	// 0000   3d 05 00 d3 5e ae 41 67 3e
	if len(data) != 4 {
		return 0.0, fmt.Errorf("could not pase float from binary")
	}
	bits := binary.LittleEndian.Uint32(data)
	float := float64(math.Float32frombits(bits))
	return math.Round(float*100) / 100, nil
}

func generateMsg(op, ln, cell int) []byte {
	addr := []byte{0xff, 0x1e, 0xc8, 0x04, byte(op), 0x04, 0x08, 0x00}
	addr[5] = byte(ln)
	addr[6] = byte(cell / 60)
	addr[7] = byte(cell % 60)
	return addr
}

func generateWriteMsg(op, ln, cell int, value []byte) []byte {
	// addr[3] seems to be 0x05 for writing
	addr := []byte{0xff, 0x1e, 0xc8, byte(4 + len(value)), byte(op), 0x04, 0x08, 0x00}
	addr[5] = byte(ln)
	addr[6] = byte(cell / 60)
	addr[7] = byte(cell % 60)
	addr = append(addr, value...)
	return addr
}

// RRP reads real segment var. (verified working).
func RRP(buf *bufio.Reader, w io.Writer, ln, cell int) (float64, error) {
	addr := generateMsg(0xb6, ln, cell)
	b, err := Send(buf, w, addr)
	if err != nil {
		return 0.0, err
	}
	return AsRoundedFloat(b.Payload())
}

// RLP reads logic segment var. (not working??? TODO).
func RLP(buf *bufio.Reader, w io.Writer, ln, cell int) (bool, error) {
	addr := generateMsg(0xb3, ln, cell)
	b, err := Send(buf, w, addr)
	if err != nil {
		return false, err
	}
	p := b.Payload()
	if len(p) != 1 {
		return false, fmt.Errorf("wrong length on RLP expected 1 byte")
	}
	if p[0] != 0 {
		return true, nil
	}
	return false, nil
}

// SRP Set real segment var.
func SRP(buf *bufio.Reader, w io.Writer, ln, cell int, val float64) error {
	addr := generateWriteMsg(0x32, ln, cell, FloatTobytes(val))
	b, err := Send(buf, w, addr)
	if err != nil {
		return err
	}
	// TODO verify this is working
	if !bytes.Equal(b, []byte{0x3d, 0x1, 0x0, 0x1, 0x3e}) {
		return fmt.Errorf("error response after SRP got: %s", PrintHex(b))
	}
	return nil
}

// RXP reads index segment var. (verified working).
func RXP(buf *bufio.Reader, w io.Writer, ln, cell int) (int, error) {
	addr := generateMsg(0x34, ln, cell)
	b, err := Send(buf, w, addr)
	if err != nil {
		return 0, err
	}
	p := b.Payload()
	if len(p) != 1 {
		return 0, fmt.Errorf("wrong length on RXP expected 1 byte")
	}
	return int(p[0]), nil
}

// SXP send index segment.
func SXP(buf *bufio.Reader, w io.Writer, ln, cell, val int) error {
	addr := generateWriteMsg(0xb0, ln, cell, []byte{byte(val)})
	b, err := Send(buf, w, addr)
	if err != nil {
		return err
	}
	if !bytes.Equal(b, []byte{0x3d, 0x1, 0x0, 0x1, 0x3e}) {
		return fmt.Errorf("error response after SXP")
	}
	return nil
}

func Send(buf *bufio.Reader, w io.Writer, data []byte) (Message, error) {
	sendMsg := compileMsg(data)
	logrus.Debugf("send msg: %x\n", sendMsg)
	if _, err := w.Write(sendMsg); err != nil {
		return nil, fmt.Errorf("error writing to exoline connection: %w", err)
	}
	return read(buf)
}

/*
 raw msghex: 0x3d 0x5 0x0 0xf5 0x1b 0xe4 0xaf 0x41 0x5 0x3e
read msghex: 0x3d 0x5 0x0 0xf5 0xaf 0x41 0x5 0x3e
error fetching data: data hex: 0x3d 0x5 0x0 0xf5 0xaf 0x41 0x5 0x3e  could not pase float from binary
*/

type Message []byte

func (m Message) Payload() []byte {
	return m[3 : len(m)-2]
}

func read(buf *bufio.Reader) (Message, error) {
	msg, err := buf.ReadBytes(0x3e)
	if err != nil {
		return nil, err
	}
	logrus.Debug("raw msg", PrintHex(msg))

	// handle escape byte
	indexesToFlip := []int{}
	n := 0
	for _, v := range msg {
		if v == 0x1b {
			indexesToFlip = append(indexesToFlip, n)
			continue
		}
		msg[n] = v
		n++
	}
	msg = msg[:n]

	for _, v := range indexesToFlip {
		msg[v] ^= 0xff
	}

	logrus.Debug("read msg", PrintHex(msg))

	return msg, nil
}

func PrintHex(data []byte) string {
	str := "hex: "
	for _, v := range data {
		str += fmt.Sprintf("%#x ", v)
	}
	return str
}

func compileMsg(b []byte) []byte {
	shouldEscape := []uint8{
		0x3c, // client start
		0x3d, // server response
		0x1b, // escape char
		0x3e, // last in msg
	}

	// calculate checksum
	var s byte
	for _, v := range b {
		s ^= v
	}

	escapeChecksum := false
	for _, b := range shouldEscape {
		if s == b {
			escapeChecksum = true
		}
	}

	// []byte{0x3c, 0xff, 0x1e, 0xc8, 0x04, 0xb6, 0x03, 0x02, 0x0c, 0x96, 0x3e}
	msg := []byte{0x3c}

	for _, a := range b {
		escape := false
		for _, b := range shouldEscape {
			if a == b {
				escape = true
			}
		}

		if escape {
			msg = append(msg, 0x1b, a^0xff) // add escape and bit invert here
		} else {
			msg = append(msg, a)
		}
	}
	if escapeChecksum {
		msg = append(msg, 0x1b, s^0xff, 0x3e) // add escape and bit invert here
		return msg
	}
	msg = append(msg, []byte{s, 0x3e}...)
	return msg
}
