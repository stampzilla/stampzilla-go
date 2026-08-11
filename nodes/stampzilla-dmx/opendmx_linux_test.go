//go:build linux

package main

import (
	"fmt"
	"os"
	"testing"
	"time"
	"unsafe"

	"golang.org/x/sys/unix"
)

func TestTermios2ForSetsCustomBaud(t *testing.T) {
	got := termios2For(250000)

	if got.Cflag&unix.BOTHER == 0 {
		t.Errorf("Cflag = %#x, want BOTHER set", got.Cflag)
	}
	if got.Ispeed != 250000 {
		t.Errorf("Ispeed = %d, want 250000", got.Ispeed)
	}
	if got.Ospeed != 250000 {
		t.Errorf("Ospeed = %d, want 250000", got.Ospeed)
	}
	if got.Cflag&unix.CS8 == 0 {
		t.Errorf("Cflag = %#x, want CS8 set (8 data bits)", got.Cflag)
	}
	if got.Cflag&unix.CSTOPB == 0 {
		t.Errorf("Cflag = %#x, want CSTOPB set (2 stop bits)", got.Cflag)
	}
}

func TestTermios2ForDifferentBaud(t *testing.T) {
	got := termios2For(9600)
	if got.Ispeed != 9600 || got.Ospeed != 9600 {
		t.Errorf("Ispeed/Ospeed = %d/%d, want 9600/9600", got.Ispeed, got.Ospeed)
	}
}

// TestOpenDMXIoctlSequenceOnPTY exercises the exact ioctl sequence
// openDMXOutput uses (TCSETS2/BOTHER, TCSBRK drain, TIOCSBRK, TIOCCBRK)
// against a real pseudo-terminal, which implements the kernel tty layer
// like a genuine serial port would. It doesn't require root or real
// hardware, and it validates the ioctl encoding is well-formed on the
// actual syscall interface - it cannot confirm real DMX electrical timing,
// which needs a scope or real fixtures.
func TestOpenDMXIoctlSequenceOnPTY(t *testing.T) {
	master, err := os.OpenFile("/dev/ptmx", unix.O_RDWR|unix.O_NOCTTY, 0)
	if err != nil {
		t.Skipf("cannot open /dev/ptmx: %v", err)
	}
	defer master.Close()

	if err := unix.IoctlSetPointerInt(int(master.Fd()), unix.TIOCSPTLCK, 0); err != nil {
		t.Skipf("unlockpt: %v", err)
	}
	n, err := unix.IoctlGetInt(int(master.Fd()), unix.TIOCGPTN)
	if err != nil {
		t.Skipf("ioctl(TIOCGPTN): %v", err)
	}

	slave, err := os.OpenFile(fmt.Sprintf("/dev/pts/%d", n), unix.O_RDWR|unix.O_NOCTTY, 0)
	if err != nil {
		t.Skipf("open pty slave: %v", err)
	}
	defer slave.Close()
	fd := slave.Fd()

	t2 := termios2For(openDMXBaud)
	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, fd, uintptr(unix.TCSETS2), uintptr(unsafe.Pointer(&t2))); errno != 0 {
		t.Errorf("TCSETS2: %v", errno)
	}
	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, fd, uintptr(unix.TCSBRK), 1); errno != 0 {
		t.Errorf("TCSBRK (drain): %v", errno)
	}
	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, fd, uintptr(unix.TIOCSBRK), 0); errno != 0 {
		t.Errorf("TIOCSBRK: %v", errno)
	}
	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, fd, uintptr(unix.TIOCCBRK), 0); errno != 0 {
		t.Errorf("TIOCCBRK: %v", errno)
	}
}

func TestOpenDMXOutputSendTiming(t *testing.T) {
	master, err := os.OpenFile("/dev/ptmx", unix.O_RDWR|unix.O_NOCTTY, 0)
	if err != nil {
		t.Skipf("cannot open /dev/ptmx: %v", err)
	}
	defer master.Close()

	if err := unix.IoctlSetPointerInt(int(master.Fd()), unix.TIOCSPTLCK, 0); err != nil {
		t.Skipf("unlockpt: %v", err)
	}
	n, err := unix.IoctlGetInt(int(master.Fd()), unix.TIOCGPTN)
	if err != nil {
		t.Skipf("ioctl(TIOCGPTN): %v", err)
	}

	slave, err := os.OpenFile(fmt.Sprintf("/dev/pts/%d", n), unix.O_RDWR|unix.O_NOCTTY, 0)
	if err != nil {
		t.Skipf("open pty slave: %v", err)
	}
	defer slave.Close()

	o := &openDMXOutput{f: slave}

	channels := make([]byte, 24) // minChannels is 24, so 24 clamped channels

	// Send first frame
	start := time.Now()
	if err := o.Send(channels); err != nil {
		t.Fatalf("first Send failed: %v", err)
	}

	// Verify lastWriteTime and lastWriteDuration are set properly
	if o.lastWriteTime.Before(start) {
		t.Errorf("lastWriteTime not updated correctly, got %v, started at %v", o.lastWriteTime, start)
	}

	// Clamped to min 24 channels + 1 start code = 25 bytes.
	// Expected duration: 25 * 44us + 200us = 1300us
	expectedDuration := 25*44*time.Microsecond + 200*time.Microsecond
	if o.lastWriteDuration != expectedDuration {
		t.Errorf("lastWriteDuration = %v, want %v", o.lastWriteDuration, expectedDuration)
	}

	// Send second frame immediately. It should wait for the first frame to complete transmission.
	send2Start := time.Now()
	if err := o.Send(channels); err != nil {
		t.Fatalf("second Send failed: %v", err)
	}
	duration := time.Since(send2Start)

	// We expect the sleep to take approximately expectedDuration.
	// Allow for some minor scheduler variations, but it should be at least a reasonable portion of expectedDuration.
	if duration < 500*time.Microsecond {
		t.Errorf("second Send completed too fast: took %v, expected sleep ~%v", duration, expectedDuration)
	}
}
