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
	mu                sync.Mutex
	f                 *os.File
	lastWriteTime     time.Time
	lastWriteDuration time.Duration
}

// openOpenDMXOutput opens the serial port used to talk to the DMX cable.
func openOpenDMXOutput(name string) (*openDMXOutput, error) {
	f, err := os.OpenFile(name, unix.O_RDWR|unix.O_NOCTTY, 0)
	if err != nil {
		return nil, err
	}

	if err := setBaudRate(f.Fd(), openDMXBaud); err != nil {
		f.Close()
		return nil, err
	}

	// Assert RTS & DTR lines (Enables RS-485 transceiver DE pin on FT232 cables)
	fd := int(f.Fd())
	lines := unix.TIOCM_RTS | unix.TIOCM_DTR
	_ = unix.IoctlSetInt(fd, unix.TIOCMBIS, lines)

	return &openDMXOutput{f: f}, nil
}

func (o *openDMXOutput) Send(channels []byte) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.f == nil {
		return os.ErrClosed
	}

	fd := o.f.Fd()

	// 1. Wait for preceding bytes in the hardware FTDI FIFO and driver buffer to finish transmitting.
	// Since tcdrain/TCSBRK on USB-serial adapters can return before the hardware FIFO is completely
	// empty, we enforce the required physical transmission time ourselves to prevent asserting
	// a BREAK in the middle of an active byte transmission.
	if !o.lastWriteTime.IsZero() {
		elapsed := time.Since(o.lastWriteTime)
		if elapsed < o.lastWriteDuration {
			time.Sleep(o.lastWriteDuration - elapsed)
		}
	}

	// 2. Change baud rate to 50,000 baud to generate the BREAK.
	// At 50,000 baud 8N2 (11 bits total per byte), transmitting 1 byte takes exactly 220 microseconds:
	// - 1 start bit (low) + 8 data bits of 0x00 (low) = 9 low bits (180us BREAK)
	// - 2 stop bits (high) = 2 high bits (40us Mark-After-Break)
	// This generates a hardware-precise, standard-compliant Break and MAB
	// that cannot overflow the decoder's firmware timers or cause line-fault bugs.
	if err := setBaudRate(fd, 50000); err != nil {
		return err
	}

	// 3. Send a single 0x00 byte to generate the Break + MAB sequence
	if _, err := o.f.Write([]byte{0x00}); err != nil {
		return err
	}

	// 4. Wait for the BREAK/MAB byte to finish transmitting physically.
	// Since the kernel buffer is tiny (1 byte), TCSBRK/drain behaves correctly.
	// We add a 250us safety sleep to ensure the hardware serializer has completely
	// shifted out the stop bits before we re-configure the baud rate generator.
	_ = ioctl(fd, uintptr(unix.TCSBRK), 1)
	preciseSleep(250 * time.Microsecond)

	// 5. Restore baud rate back to 250,000 baud for standard DMX data packet transmission.
	if err := setBaudRate(fd, openDMXBaud); err != nil {
		return err
	}

	data := clampChannels(channels)
	frame := make([]byte, 0, len(data)+1)
	frame = append(frame, dmxStartCode)
	frame = append(frame, data...)

	_, err := o.f.Write(frame)
	if err == nil {
		o.lastWriteTime = time.Now()
		// At 250,000 baud 8N2 (11 bits total per byte), transmitting 1 byte takes exactly 44 microseconds.
		// We add a 200 microseconds safety margin to be absolutely certain the chip's transmitter
		// is completely done shifting out the last stop bit before the next loop tries to assert BREAK.
		o.lastWriteDuration = time.Duration(len(frame))*44*time.Microsecond + 200*time.Microsecond
	}
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

func preciseSleep(d time.Duration) {
	start := time.Now()
	for time.Since(start) < d {
	}
}

func setBaudRate(fd uintptr, baud uint32) error {
	t := termios2For(baud)
	if errno := ioctl(fd, uintptr(unix.TCSETS2), uintptr(unsafe.Pointer(&t))); errno != 0 {
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
