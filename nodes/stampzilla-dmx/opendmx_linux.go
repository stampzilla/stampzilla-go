//go:build linux

package main

import (
	"os"
	"sync"
	"time"
	"unsafe"

	"golang.org/x/sys/unix"
)

// Open DMX USB-style adapters (a bare FTDI/RS485 cable with no onboard
// microcontroller, unlike an Enttec DMX USB Pro-class widget) have no
// firmware to generate DMX-512 line timing or interpret a framing protocol.
// The host has to do it: hold the line low for a break, release it for a
// mark-after-break, then send the start code and channel bytes at a fixed
// 250,000 baud, 8 data bits, 2 stop bits, no parity.
//
// This is software-timed: time.Sleep at the ~100us scale has real jitter on
// a general-purpose Linux scheduler, and the FTDI kernel driver's default
// 16ms USB latency timer can add per-write lag. Most DMX receivers tolerate
// this fine, but it is inherently less rock-solid than a firmware-timed
// widget. See README.md for a latency_timer tuning tip.
const (
	openDMXBaud = 250000 // DMX-512's fixed line rate; not configurable

	breakDuration  = 176 * time.Microsecond // DMX-512 minimum is 92us
	markAfterBreak = 20 * time.Microsecond  // DMX-512 minimum is 12us
)

// termios2For builds the termios value used to configure the serial port at
// an arbitrary baud rate via the Linux-specific BOTHER/TCSETS2 mechanism.
// golang.org/x/sys/unix's Termios struct already carries the kernel's
// termios2 layout (trailing Ispeed/Ospeed fields), so no custom struct is
// needed - only the request (TCSETS2) and flag (BOTHER) differ from the
// standard termios path tarm/serial itself uses.
func termios2For(baud uint32) unix.Termios {
	t := unix.Termios{
		Iflag:  unix.IGNPAR,
		Cflag:  unix.CS8 | unix.CSTOPB | unix.CREAD | unix.CLOCAL | unix.BOTHER,
		Ispeed: baud,
		Ospeed: baud,
	}
	t.Cc[unix.VMIN] = 0
	t.Cc[unix.VTIME] = 0
	return t
}

// openDMXOutput writes DMX frames to a raw Open DMX USB-style serial widget.
type openDMXOutput struct {
	mu sync.Mutex
	f  *os.File
}

// openOpenDMXOutput opens the serial port used to talk to the DMX cable.
func openOpenDMXOutput(name string) (*openDMXOutput, error) {
	f, err := os.OpenFile(name, unix.O_RDWR|unix.O_NOCTTY, 0)
	if err != nil {
		return nil, err
	}

	t := termios2For(openDMXBaud)
	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, f.Fd(), uintptr(unix.TCSETS2), uintptr(unsafe.Pointer(&t))); errno != 0 {
		f.Close()
		return nil, errno
	}

	return &openDMXOutput{f: f}, nil
}

func (o *openDMXOutput) Send(channels []byte) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	fd := o.f.Fd()

	// Drain (wait for the previous frame to finish transmitting) without
	// sending a break, so we never assert a break mid-byte and corrupt the
	// tail of the prior frame. A nonzero argument means drain-only on Linux.
	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, fd, uintptr(unix.TCSBRK), 1); errno != 0 {
		return errno
	}

	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, fd, uintptr(unix.TIOCSBRK), 0); errno != 0 {
		return errno
	}
	time.Sleep(breakDuration)
	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, fd, uintptr(unix.TIOCCBRK), 0); errno != 0 {
		return errno
	}
	time.Sleep(markAfterBreak)

	data := clampChannels(channels)
	frame := make([]byte, 0, len(data)+1)
	frame = append(frame, dmxStartCode)
	frame = append(frame, data...)

	_, err := o.f.Write(frame)
	return err
}

func (o *openDMXOutput) Close() error {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.f == nil {
		return nil
	}
	err := o.f.Close()
	o.f = nil
	return err
}
