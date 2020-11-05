package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"math"

	"github.com/sirupsen/logrus"
)

func asRoundedFloat(data []byte) (float64, error) {
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

func Send(buf *bufio.Reader, w io.Writer, data []byte) ([]byte, error) {
	sendMsg := compileMsg(data)
	logrus.Debugf("send msg: %x\n", sendMsg)
	_, err := w.Write(sendMsg)
	if err != nil {
		logrus.Error(err)
	}
	return read(buf)
}

/*
 raw msghex: 0x3d 0x5 0x0 0xf5 0x1b 0xe4 0xaf 0x41 0x5 0x3e
read msghex: 0x3d 0x5 0x0 0xf5 0xaf 0x41 0x5 0x3e
error fetching data: data hex: 0x3d 0x5 0x0 0xf5 0xaf 0x41 0x5 0x3e  could not pase float from binary
*/

func read(buf *bufio.Reader) ([]byte, error) {
	msg, err := buf.ReadBytes(0x3e)
	if err != nil {
		return nil, err
	}
	logrus.Debug("raw msg", printHex(msg))

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

	logrus.Debug("read msg", printHex(msg))

	return msg, nil
}

func printHex(data []byte) string {
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

	//[]byte{0x3c, 0xff, 0x1e, 0xc8, 0x04, 0xb6, 0x03, 0x02, 0x0c, 0x96, 0x3e}
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
