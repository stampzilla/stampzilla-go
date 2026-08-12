package main

import (
	"flag"
	"fmt"
	"io"
	"strings"
	"time"
)

// This file implements a standalone DMX bring-up test that talks directly to
// openOpenDMXOutput (see selftest_linux.go), bypassing the stampzilla
// server, config push and device state entirely. It exists because a wedged
// decoder (address buttons locked up, front panel unreadable) can't be
// diagnosed by staring at config or logs alone - it needs something driving
// known channel values onto the wire on demand, with nothing else in the
// way.
//
// It is invoked as a subcommand (`stampzilla-dmx selftest ...`), not as
// flags on the main command line. The node's own flags (-host, -port,
// -loglevel, ...) are parsed by github.com/koding/multiconfig from inside
// node.New(), using a *private* flag.FlagSet that has no way to learn about
// extra flags - so a parallel top-level `-selftest-*` flag either gets
// rejected by multiconfig as unknown, or (if it runs first, as it used to)
// silently swallows every real flag, including --help and -version. A
// leading positional "selftest" token doesn't have that problem: stdlib
// flag.Parse stops at the first non-flag argument without erroring, so it's
// invisible to multiconfig either way, and it's consumed by this package's
// own dispatch in main() before node.New() ever runs.
//
// Usage:
//
//	stampzilla-dmx selftest -port /dev/ttyUSB0 -mode full
//	stampzilla-dmx selftest -port /dev/ttyUSB0 -mode walk
//	stampzilla-dmx selftest -port /dev/ttyUSB0 -mode ramp -break-mode ioctl -de-mode assert

const selftestSubcommand = "selftest"

// selftestConfig holds the parsed selftest subcommand flags.
type selftestConfig struct {
	port      string
	mode      string
	channels  int
	fps       int
	breakMode breakMode
	deMode    deMode
	echo      bool
	walkHold  time.Duration
}

// selftestRequested reports whether args (== os.Args[1:]) invoke the
// selftest subcommand.
func selftestRequested(args []string) bool {
	return len(args) > 0 && args[0] == selftestSubcommand
}

// legacySelftestHint detects the old `-selftest-*` top-level flag syntax
// (replaced by the `selftest` subcommand, which fixed those flags breaking
// every normal node flag) and returns a hint message, or "" if args don't
// look like the old syntax.
func legacySelftestHint(args []string) string {
	for _, a := range args {
		if strings.HasPrefix(a, "-selftest") || strings.HasPrefix(a, "--selftest") {
			return "dmx: -selftest-* flags were replaced by a subcommand - run `stampzilla-dmx selftest -h`"
		}
	}
	return ""
}

// parseSelftestArgs parses the flags following the "selftest" subcommand
// (i.e. os.Args[2:]). out receives usage/error text: os.Stderr in
// production, a buffer in tests.
func parseSelftestArgs(args []string, out io.Writer) (*selftestConfig, error) {
	fs := flag.NewFlagSet(selftestSubcommand, flag.ContinueOnError)
	fs.SetOutput(out)
	fs.Usage = func() {
		fmt.Fprintln(out, "Usage: stampzilla-dmx selftest -port <device> [flags]")
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Drives a DMX cable directly, bypassing the stampzilla server, to bring up")
		fmt.Fprintln(out, "or diagnose hardware. The node's own flags (-host, -port, -loglevel, ...)")
		fmt.Fprintln(out, "are shown by running stampzilla-dmx without a subcommand.")
		fmt.Fprintln(out)
		fs.PrintDefaults()
	}

	port := fs.String("port", "", "serial device to drive, e.g. /dev/ttyUSB0 (required)")
	mode := fs.String("mode", "full", `pattern: "full" (all channels at 255), "walk" (light one channel at a time), or "ramp" (slow fade on all channels)`)
	channels := fs.Int("channels", minChannels, "number of DMX channels to drive")
	fps := fs.Int("fps", defaultFPS, "frames per second to resend at")
	breakModeFlag := fs.String("break-mode", string(breakModeBaud), `break generation: "baud" or "ioctl"`)
	deModeFlag := fs.String("de-mode", string(deModeNone), `RS485 driver-enable handling: "none", "assert" or "clear"`)
	echo := fs.Bool("echo", false, "also read from the port while transmitting, logging any bytes seen (best-effort - most adapters won't echo their own output)")
	walkHold := fs.Duration("walk-hold", 2*time.Second, "how long to hold each channel high in walk mode")

	if err := fs.Parse(args); err != nil {
		// flag.Parse already printed the error and usage to out.
		return nil, err
	}

	invalid := func(format string, a ...any) (*selftestConfig, error) {
		fs.Usage()
		return nil, fmt.Errorf(format, a...)
	}

	if *port == "" {
		return invalid("selftest: -port is required")
	}
	if *channels < 1 {
		return invalid("selftest: -channels must be >= 1")
	}
	if *fps < 1 {
		return invalid("selftest: -fps must be >= 1")
	}
	switch *mode {
	case "full", "walk", "ramp":
	default:
		return invalid("selftest: unknown -mode %q (want full, walk or ramp)", *mode)
	}
	bm := breakMode(*breakModeFlag)
	switch bm {
	case breakModeBaud, breakModeIoctl:
	default:
		return invalid("selftest: unknown -break-mode %q (want %q or %q)", *breakModeFlag, breakModeBaud, breakModeIoctl)
	}
	dm := deMode(*deModeFlag)
	switch dm {
	case deModeNone, deModeAssert, deModeClear:
	default:
		return invalid("selftest: unknown -de-mode %q (want %q, %q or %q)", *deModeFlag, deModeNone, deModeAssert, deModeClear)
	}

	return &selftestConfig{
		port:      *port,
		mode:      *mode,
		channels:  *channels,
		fps:       *fps,
		breakMode: bm,
		deMode:    dm,
		echo:      *echo,
		walkHold:  *walkHold,
	}, nil
}

// selftestFrame computes one frame's channel values for cfg.mode at the
// given elapsed time since the selftest started. It has no side effects
// (no I/O, no logging), so it is directly unit-testable without a real or
// fake serial port.
func selftestFrame(cfg *selftestConfig, elapsed time.Duration) []byte {
	frame := make([]byte, cfg.channels)
	switch cfg.mode {
	case "full":
		for i := range frame {
			frame[i] = 255
		}

	case "ramp":
		const period = 2 * time.Second
		e := elapsed % period
		phase := float64(e) / float64(period) // 0..1, one full up/down cycle per period
		level := phase * 2
		if level > 1 {
			level = 2 - level
		}
		v := byte(level * 255)
		for i := range frame {
			frame[i] = v
		}

	case "walk":
		frame[walkChannel(cfg, elapsed)] = 255
	}
	return frame
}

// walkChannel returns which 0-based channel index "walk" mode should be
// driving at the given elapsed time.
func walkChannel(cfg *selftestConfig, elapsed time.Duration) int {
	hold := cfg.walkHold
	if hold <= 0 {
		hold = 2 * time.Second
	}
	return int(elapsed/hold) % cfg.channels
}
