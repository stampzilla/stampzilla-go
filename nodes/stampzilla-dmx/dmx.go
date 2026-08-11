package main

import (
	"github.com/sirupsen/logrus"
)

const (
	dmxStartCode = 0x00 // NULL START Code, as defined by the DMX-512 standard

	minChannels = 24  // pad short frames so real fixtures see a stable, full-ish refresh
	maxChannels = 512 // a DMX-512 universe has 512 channels
)

// clampChannels pads/truncates channels to [minChannels, maxChannels].
func clampChannels(channels []byte) []byte {
	n := len(channels)
	if n < minChannels {
		n = minChannels
	}
	if n > maxChannels {
		n = maxChannels
	}
	if n == len(channels) {
		return channels
	}
	out := make([]byte, n)
	copy(out, channels)
	return out
}

// dmxOutput sends a full universe frame to a DMX widget.
type dmxOutput interface {
	Send(channels []byte) error
	Close() error
}

// logOutput logs frames instead of writing to hardware. It is used when no
// serial port is configured, so the node is fully exercisable without a
// physical widget attached.
type logOutput struct{}

func (logOutput) Send(channels []byte) error {
	logrus.Debugf("dmx: frame (no port configured): % x", channels)
	return nil
}

func (logOutput) Close() error { return nil }
