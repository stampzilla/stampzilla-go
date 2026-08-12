//go:build linux

package main

import (
	"fmt"
	"os"
	"sync"
	"time"
	"unsafe"

	"github.com/sirupsen/logrus"
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

	ioctlBreakDuration  = 176 * time.Microsecond // DMX-512 minimum is 92us
	ioctlMarkAfterBreak = 20 * time.Microsecond   // DMX-512 minimum is 12us
)

// termios2For builds the termios value used to configure the serial port at
// an arbitrary baud rate via the Linux-specific BOTHER/TCSETS2 mechanism.
// golang.org/x/sys/unix's Termios struct already carries the kernel's
// termios2 layout (trailing Ispeed/Ospeed fields), so no custom struct is
// needed - only the request (TCSETS2/TCSETSW2) and flag (BOTHER) differ from
// the standard termios path tarm/serial itself uses.
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
	mu                sync.Mutex
	f                 *os.File
	breakMode         breakMode
	lastWriteTime     time.Time
	lastWriteDuration time.Duration
}

// openOpenDMXOutput opens the serial port used to talk to the DMX cable.
func openOpenDMXOutput(name string, bm breakMode, dm deMode) (*openDMXOutput, error) {
	f, err := os.OpenFile(name, unix.O_RDWR|unix.O_NOCTTY, 0)
	if err != nil {
		return nil, err
	}

	if err := setBaudRate(f.Fd(), openDMXBaud); err != nil {
		f.Close()
		return nil, err
	}

	// Assert/clear RTS & DTR if configured (some RS485 cables wire the
	// transceiver's DE pin to one of these). This is best-effort: plenty of
	// adapters (and every pty used in tests) legitimately reject TIOCMBIS/
	// TIOCMBIC, so a failure here is logged, not fatal.
	if dm != deModeNone && dm != "" {
		fd := int(f.Fd())
		lines := unix.TIOCM_RTS | unix.TIOCM_DTR
		req := uint(unix.TIOCMBIS)
		if dm == deModeClear {
			req = uint(unix.TIOCMBIC)
		}
		if err := unix.IoctlSetInt(fd, req, lines); err != nil {
			logrus.Warnf("dmx: set RTS/DTR (mode=%s): %s", dm, err)
		}
	}

	if bm == "" {
		bm = breakModeBaud
	}
	return &openDMXOutput{f: f, breakMode: bm}, nil
}

func (o *openDMXOutput) Send(channels []byte) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.f == nil {
		return os.ErrClosed
	}

	fd := o.f.Fd()

	// Wait for the previous frame to finish transmitting before touching the
	// line again. See sendBreakBaud/sendBreakIoctl for how each break mode
	// protects against reconfiguring the port mid-byte.
	if !o.lastWriteTime.IsZero() {
		elapsed := time.Since(o.lastWriteTime)
		if elapsed < o.lastWriteDuration {
			time.Sleep(o.lastWriteDuration - elapsed)
		}
	}

	var err error
	switch o.breakMode {
	case breakModeIoctl:
		err = sendBreakIoctl(fd)
	default:
		err = sendBreakBaud(fd, o.f)
	}
	if err != nil {
		return err
	}

	data := clampChannels(channels)
	frame := make([]byte, 0, len(data)+1)
	frame = append(frame, dmxStartCode)
	frame = append(frame, data...)

	_, err = o.f.Write(frame)
	if err == nil {
		o.lastWriteTime = time.Now()
		// At 250,000 baud 8N2 (11 bits total per byte), transmitting 1 byte takes exactly 44 microseconds.
		// We add a 200 microseconds safety margin to be absolutely certain the chip's transmitter
		// is completely done shifting out the last stop bit before the next loop tries to assert BREAK.
		o.lastWriteDuration = time.Duration(len(frame))*44*time.Microsecond + 200*time.Microsecond
	}
	return err
}

// sendBreakBaud generates the BREAK+MAB by dropping to 50,000 baud and
// writing a single 0x00 byte: at 50000 8N2 that byte is 9 low bits (180us
// break) followed by 2 high stop bits (40us MAB).
//
// Both the drop to 50000 and the restore to 250000 use TCSETSW2 (drain-then-
// set), not TCSETS2 (set-immediately). TCSETS2 applies the new divisor to
// the UART the instant the ioctl returns, with no guarantee the previous
// write has finished shifting out - on a USB-serial adapter, where the
// "hardware" FIFO is really queued inside a URB, the ioctl can return well
// before the wire is actually idle. Reconfiguring the baud generator while a
// byte is still shifting out corrupts it (and DMX receivers, including the
// address-button lockup reported against this cable, react very badly to a
// corrupted frame tail). TCSETSW2 drains first, so the divisor only changes
// once the previous byte has actually left the chip.
func sendBreakBaud(fd uintptr, f *os.File) error {
	if err := setBaudRateDrain(fd, 50000); err != nil {
		return fmt.Errorf("dmx: set break baud: %w", err)
	}

	if _, err := f.Write([]byte{0x00}); err != nil {
		return fmt.Errorf("dmx: write break byte: %w", err)
	}

	// No explicit drain here: setBaudRateDrain uses TCSETSW2, which already
	// waits for the break byte above to finish transmitting before applying
	// the new baud rate.
	if err := setBaudRateDrain(fd, openDMXBaud); err != nil {
		return fmt.Errorf("dmx: restore data baud: %w", err)
	}
	return nil
}

// sendBreakIoctl generates the BREAK+MAB with TIOCSBRK/TIOCCBRK, holding the
// line low for ioctlBreakDuration and high for ioctlMarkAfterBreak. This is
// the historical mechanism used before the baud-rate trick and is kept as a
// runtime-selectable fallback, since ftdi_sio break support differs across
// kernel versions and adapter clones.
func sendBreakIoctl(fd uintptr) error {
	if err := ioctl(fd, uintptr(unix.TCSBRK), 1); err != 0 {
		return fmt.Errorf("dmx: drain before break: %w", err)
	}
	if err := ioctl(fd, uintptr(unix.TIOCSBRK), 0); err != 0 {
		return fmt.Errorf("dmx: assert break: %w", err)
	}
	time.Sleep(ioctlBreakDuration)
	if err := ioctl(fd, uintptr(unix.TIOCCBRK), 0); err != 0 {
		return fmt.Errorf("dmx: clear break: %w", err)
	}
	time.Sleep(ioctlMarkAfterBreak)
	return nil
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

// setBaudRate applies a termios2 baud-rate change immediately (TCSETS2,
// equivalent to TCSANOW - no drain).
func setBaudRate(fd uintptr, baud uint32) error {
	t := termios2For(baud)
	if errno := ioctl(fd, uintptr(unix.TCSETS2), uintptr(unsafe.Pointer(&t))); errno != 0 {
		return errno
	}
	return nil
}

// setBaudRateDrain applies a termios2 baud-rate change only after all
// pending output has been transmitted (TCSETSW2, equivalent to TCSADRAIN).
// Used inside Send() so a baud switch can never land mid-byte.
func setBaudRateDrain(fd uintptr, baud uint32) error {
	t := termios2For(baud)
	if errno := ioctl(fd, uintptr(unix.TCSETSW2), uintptr(unsafe.Pointer(&t))); errno != 0 {
		return errno
	}
	return nil
}

func ioctl(fd uintptr, req uintptr, arg uintptr) unix.Errno {
	for {
		_, _, errno := unix.Syscall(unix.SYS_IOCTL, fd, req, arg)
		if errno != unix.EINTR {
			return errno
		}
	}
}
