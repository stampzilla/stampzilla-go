//go:build linux

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

// This file implements a standalone DMX bring-up test that talks directly to
// openOpenDMXOutput, bypassing the stampzilla server, config push and device
// state entirely. It exists because a wedged decoder (address buttons locked
// up, front panel unreadable) can't be diagnosed by staring at config or
// logs alone - it needs something driving known channel values onto the
// wire on demand, with nothing else in the way.
//
// Usage:
//
//	stampzilla-dmx -selftest-port /dev/ttyUSB0 -selftest-mode full
//	stampzilla-dmx -selftest-port /dev/ttyUSB0 -selftest-mode walk
//	stampzilla-dmx -selftest-port /dev/ttyUSB0 -selftest-mode ramp -selftest-break-mode ioctl -selftest-de-mode assert

// selftestConfig holds the parsed -selftest-* flags.
type selftestConfig struct {
	port      string
	mode      string
	channels  int
	fps       int
	breakMode string
	deMode    string
	echo      bool
	walkHold  time.Duration
}

// maybeRunSelftest checks args for -selftest-port and, if present, runs a
// selftest and returns (true, exitCode) - main should exit immediately with
// that code rather than starting the node. If -selftest-port was not given,
// it returns (false, 0) and normal node startup should proceed.
func maybeRunSelftest(args []string) (bool, int) {
	fs := flag.NewFlagSet("selftest", flag.ContinueOnError)
	port := fs.String("selftest-port", "", "run a standalone DMX selftest against this serial port instead of connecting to the stampzilla server")
	mode := fs.String("selftest-mode", "full", `selftest pattern: "full" (all channels at 255), "walk" (light one channel at a time), or "ramp" (slow fade on all channels)`)
	channels := fs.Int("selftest-channels", minChannels, "number of DMX channels to drive")
	fps := fs.Int("selftest-fps", defaultFPS, "frames per second to resend at")
	breakModeFlag := fs.String("selftest-break-mode", string(breakModeBaud), `break generation: "baud" or "ioctl"`)
	deModeFlag := fs.String("selftest-de-mode", string(deModeNone), `RS485 driver-enable handling: "none", "assert" or "clear"`)
	echo := fs.Bool("selftest-echo", false, "also read from the port while transmitting, logging any bytes seen (best-effort - most adapters won't echo their own output)")
	walkHold := fs.Duration("selftest-walk-hold", 2*time.Second, "how long to hold each channel high in walk mode")

	if err := fs.Parse(args); err != nil {
		// flag.ContinueOnError already printed the error/usage.
		if err == flag.ErrHelp {
			return true, 0
		}
		return true, 2
	}

	if *port == "" {
		return false, 0
	}

	cfg := selftestConfig{
		port:      *port,
		mode:      *mode,
		channels:  *channels,
		fps:       *fps,
		breakMode: *breakModeFlag,
		deMode:    *deModeFlag,
		echo:      *echo,
		walkHold:  *walkHold,
	}

	ctx, cancel := context.WithCancel(context.Background())
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sig
		logrus.Info("dmx: selftest: shutting down")
		cancel()
	}()

	if err := runSelftest(ctx, cfg); err != nil {
		logrus.Errorf("dmx: selftest: %s", err)
		return true, 1
	}
	return true, 0
}

func runSelftest(ctx context.Context, cfg selftestConfig) error {
	if cfg.channels < 1 {
		return fmt.Errorf("selftest-channels must be >= 1")
	}
	if cfg.fps < 1 {
		return fmt.Errorf("selftest-fps must be >= 1")
	}
	switch cfg.mode {
	case "full", "walk", "ramp":
	default:
		return fmt.Errorf("unknown selftest-mode %q (want full, walk or ramp)", cfg.mode)
	}

	out, err := openOpenDMXOutput(cfg.port, breakMode(cfg.breakMode), deMode(cfg.deMode))
	if err != nil {
		return fmt.Errorf("open %s: %w", cfg.port, err)
	}
	defer out.Close()

	logrus.Infof("dmx: selftest: port=%s mode=%s channels=%d fps=%d breakMode=%s deMode=%s echo=%v",
		cfg.port, cfg.mode, cfg.channels, cfg.fps, cfg.breakMode, cfg.deMode, cfg.echo)
	logrus.Info("dmx: selftest: press Ctrl+C to stop")

	if cfg.echo {
		go selftestEcho(ctx, out)
	}

	return selftestRun(ctx, out, cfg)
}

// selftestRun resends a frame at cfg.fps until ctx is cancelled, with frame
// content depending on cfg.mode.
func selftestRun(ctx context.Context, out *openDMXOutput, cfg selftestConfig) error {
	interval := time.Second / time.Duration(cfg.fps)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	start := time.Now()
	lastStatusLog := time.Time{}
	lastWalkChannel := -1

	for {
		select {
		case <-ctx.Done():
			return nil
		case now := <-ticker.C:
			frame := make([]byte, cfg.channels)
			switch cfg.mode {
			case "full":
				for i := range frame {
					frame[i] = 255
				}
				if time.Since(lastStatusLog) >= 2*time.Second {
					lastStatusLog = now
					logrus.Infof("dmx: selftest: sending all %d channels at 255 - check the fixture now", cfg.channels)
				}

			case "ramp":
				const period = 2 * time.Second
				elapsed := now.Sub(start) % period
				phase := float64(elapsed) / float64(period) // 0..1, one full up/down cycle per period
				level := phase * 2
				if level > 1 {
					level = 2 - level
				}
				v := byte(level * 255)
				for i := range frame {
					frame[i] = v
				}

			case "walk":
				idx := int(now.Sub(start)/cfg.walkHold) % cfg.channels
				frame[idx] = 255
				if idx != lastWalkChannel {
					lastWalkChannel = idx
					logrus.Infof("dmx: selftest: walk: DMX slot %d (address %d) ON for %s", idx+1, idx+1, cfg.walkHold)
				}
			}

			if err := out.Send(frame); err != nil {
				return fmt.Errorf("send frame: %w", err)
			}
		}
	}
}

// selftestEcho reads from the port while a frame loop is transmitting on it
// and logs whatever bytes arrive. Many bare RS485 cables leave the receiver
// enabled and will see their own transmitted bytes on the shared A/B pair;
// seeing our own start code (0x00) and slot bytes come back confirms the
// baud rate/framing made it onto the wire at all. This is best-effort only:
// a half-duplex transceiver with the receiver gated off, or one where DE/RE
// are tied together, will see nothing here even when transmission is fine.
func selftestEcho(ctx context.Context, o *openDMXOutput) {
	buf := make([]byte, 256)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		o.mu.Lock()
		f := o.f
		o.mu.Unlock()
		if f == nil {
			return
		}

		_ = f.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
		n, err := f.Read(buf)
		if err != nil {
			if os.IsTimeout(err) {
				continue
			}
			return
		}
		if n > 0 {
			logrus.Infof("dmx: selftest: echo: read %d bytes: % x", n, buf[:n])
		}
	}
}
