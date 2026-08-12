package main

import (
	"github.com/sirupsen/logrus"
)

const (
	dmxStartCode = 0x00 // NULL START Code, as defined by the DMX-512 standard

	minChannels = 24  // pad short frames so real fixtures see a stable, full-ish refresh
	maxChannels = 512 // a DMX-512 universe has 512 channels
)

// breakMode selects how openDMXOutput.Send generates the BREAK/MAB. Both are
// legitimate techniques for a bare FTDI/RS485 cable and are known to behave
// differently across ftdi_sio kernel versions and adapter clones, so it is
// exposed as a config knob (see config.go) rather than hardcoded, to let a
// stuck decoder be diagnosed by flipping a setting instead of recompiling.
// Defined here (not opendmx_linux.go) so it is available to config.go and
// opendmx_other.go on every platform, not just linux.
type breakMode string

const (
	// breakModeBaud drops the baud rate to 50000 and writes a single 0x00
	// byte, which produces a hardware-precise 180us break + 40us MAB.
	breakModeBaud breakMode = "baud"
	// breakModeIoctl uses TIOCSBRK/TIOCCBRK to assert and clear a break
	// condition directly, with fixed sleeps in between.
	breakModeIoctl breakMode = "ioctl"
)

// deMode selects whether openOpenDMXOutput touches the RTS/DTR modem control
// lines when opening the port. Some FT232-based RS485 cables wire the
// transceiver's driver-enable (DE) pin to RTS or DTR; whether asserting or
// clearing that pin enables the driver (or does nothing) is adapter-specific
// and cannot be discovered without a scope, so it is a runtime choice rather
// than a fixed behavior.
type deMode string

const (
	deModeNone   deMode = "none"   // never touch RTS/DTR (matches plain adapters)
	deModeAssert deMode = "assert" // assert RTS+DTR (drive the pins low on FT232)
	deModeClear  deMode = "clear"  // clear RTS+DTR (drive the pins high on FT232)
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
