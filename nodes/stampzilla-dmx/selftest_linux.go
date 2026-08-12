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

// runSelftestCommand is main()'s entry point for `stampzilla-dmx selftest
// ...` on Linux, the only platform openOpenDMXOutput is implemented on. args
// is os.Args[2:] (everything after the "selftest" token). The return value
// is the process exit code.
func runSelftestCommand(args []string) int {
	cfg, err := parseSelftestArgs(args, os.Stderr)
	if err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}

	out, err := openOpenDMXOutput(cfg.port, cfg.breakMode, cfg.deMode)
	if err != nil {
		logrus.Errorf("dmx: selftest: open %s: %s", cfg.port, err)
		return 1
	}
	defer out.Close()

	logrus.Infof("dmx: selftest: port=%s mode=%s channels=%d fps=%d breakMode=%s deMode=%s echo=%v",
		cfg.port, cfg.mode, cfg.channels, cfg.fps, cfg.breakMode, cfg.deMode, cfg.echo)
	logrus.Info("dmx: selftest: press Ctrl+C to stop")

	ctx, cancel := context.WithCancel(context.Background())
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sig
		logrus.Info("dmx: selftest: shutting down")
		cancel()
	}()

	if cfg.echo {
		go selftestEcho(ctx, out)
	}

	if err := selftestRun(ctx, out, cfg); err != nil {
		logrus.Errorf("dmx: selftest: %s", err)
		return 1
	}
	return 0
}

// selftestRun resends a frame (built by the pure selftestFrame) at cfg.fps
// until ctx is cancelled, logging progress for the "full" and "walk" modes.
func selftestRun(ctx context.Context, out *openDMXOutput, cfg *selftestConfig) error {
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
			elapsed := now.Sub(start)
			frame := selftestFrame(cfg, elapsed)

			switch cfg.mode {
			case "full":
				if time.Since(lastStatusLog) >= 2*time.Second {
					lastStatusLog = now
					logrus.Infof("dmx: selftest: sending all %d channels at 255 - check the fixture now", cfg.channels)
				}
			case "walk":
				if idx := walkChannel(cfg, elapsed); idx != lastWalkChannel {
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
