package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/models/devices"
	"github.com/stampzilla/stampzilla-go/v2/pkg/node"
)

const (
	defaultPort = "23"
	deviceID    = "1"
)

func main() {
	n, a := start()
	if n == nil {
		return
	}
	n.Wait()
	a.close()
}

// avr holds the live TCP connection to the Denon receiver and serialises writes.
type avr struct {
	mu   sync.Mutex
	conn net.Conn
}

// send writes cmd followed by a CR to the receiver.
func (a *avr) send(cmd string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.conn == nil {
		return fmt.Errorf("denonavr: not connected")
	}

	if err := a.conn.SetWriteDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return err
	}
	_, err := fmt.Fprintf(a.conn, "%s\r", cmd)
	_ = a.conn.SetWriteDeadline(time.Time{})
	if err != nil {
		return err
	}

	time.Sleep(50 * time.Millisecond)
	return nil
}

// set replaces the active connection, closing any previous one.
// Pass nil to clear the connection without opening a new one.
func (a *avr) set(conn net.Conn) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.conn != nil {
		a.conn.Close()
	}
	a.conn = conn
}

// close terminates the active connection.
func (a *avr) close() {
	a.set(nil)
}

// Config is the per-node application config pushed from the stampzilla server.
type Config struct {
	Host string `json:"host"`
}

var nodeConfig = &Config{}

func updatedConfig(connectToHost chan<- string) node.OnFunc {
	return func(data json.RawMessage) error {
		newConf := &Config{}
		if err := json.Unmarshal(data, newConf); err != nil {
			return fmt.Errorf("denonavr: bad config: %w", err)
		}
		if newConf.Host != nodeConfig.Host {
			logrus.Infof("denonavr: new host from config: %s", newConf.Host)
			nodeConfig = newConf
			connectToHost <- newConf.Host
		}
		return nil
	}
}

func start() (*node.Node, *avr) {
	connectToHost := make(chan string, 1)
	a := &avr{}

	n := node.New("denonavr")
	n.OnConfig(updatedConfig(connectToHost))

	if err := n.Connect(); err != nil {
		logrus.Error(err)
		return nil, nil
	}

	ensureDevice(n)

	n.OnRequestStateChange(func(state devices.State, _ *devices.Device) error {
		var err error

		state.Bool("on", func(on bool) {
			if on {
				err = a.send("PWON")
			} else {
				err = a.send("PWSTANDBY")
			}
		})
		if err != nil {
			return err
		}

		state.Float("volume", func(v float64) {
			err = a.send(volumeToMV(v))
		})
		if err != nil {
			return err
		}

		state.String("source", func(s string) {
			err = a.send("SI" + s)
		})
		return err
	})

	go connectionWorker(n, a, connectToHost)

	return n, a
}

// ensureDevice registers the Denon AVR device with the node if it does not exist yet.
func ensureDevice(n *node.Node) {
	if n.GetDevice(deviceID) != nil {
		return
	}
	n.AddOrUpdate(&devices.Device{
		Type:   "mediaplayer",
		ID:     devices.ID{ID: deviceID},
		Name:   "Denon AVR",
		Online: false,
		Traits: []string{"OnOff", "Volume"},
		State:  devices.State{"on": false, "volume": 0.0, "source": ""},
	})
}

// normaliseHost appends the default telnet port (23) when none is present.
func normaliseHost(host string) string {
	if _, _, err := net.SplitHostPort(host); err != nil {
		return net.JoinHostPort(host, defaultPort)
	}
	return host
}

// connectionWorker maintains a persistent TCP connection to the Denon receiver.
// It reconnects automatically on network errors and reacts to host changes
// delivered via connectToHost (populated by OnConfig).
func connectionWorker(n *node.Node, a *avr, connectToHost <-chan string) {
	var host string

	// Wait for the first host before attempting any connection.
	select {
	case h := <-connectToHost:
		host = h
	case <-n.Stopped():
		return
	}

	for {
		if strings.TrimSpace(host) == "" {
			n.SetDeviceOnline(deviceID, false)
			select {
			case h := <-connectToHost:
				host = h
				continue
			case <-n.Stopped():
				return
			}
		}

		addr := normaliseHost(host)
		logrus.Infof("denonavr: connecting to %s", addr)

		dialCtx, dialCancel := context.WithTimeout(context.Background(), 10*time.Second)
		conn, err := (&net.Dialer{}).DialContext(dialCtx, "tcp", addr)
		dialCancel()

		if err != nil {
			logrus.Errorf("denonavr: dial %s: %s", addr, err)
			n.SetDeviceOnline(deviceID, false)
			select {
			case h := <-connectToHost:
				host = h
			case <-time.After(10 * time.Second):
			case <-n.Stopped():
				return
			}
			continue
		}

		logrus.Infof("denonavr: connected to %s", addr)
		a.set(conn)
		ensureDevice(n)
		n.SetDeviceOnline(deviceID, true)

		// Query initial state.
		_ = a.send("PW?")
		_ = a.send("MV?")
		_ = a.send("SI?")

		// Start an async reader that feeds incoming events into device state.
		disconnected := make(chan struct{})
		go func() {
			defer close(disconnected)
			readLoop(n, conn)
		}()

		// Block until the link drops, the target host changes, or we shut down.
		select {
		case <-n.Stopped():
			a.close() // triggers readLoop to exit with an I/O error
			<-disconnected
			return

		case h := <-connectToHost:
			host = h
			n.SetDeviceOnline(deviceID, false)
			a.close() // triggers readLoop to exit
			<-disconnected
			// fall through to reconnect with new host

		case <-disconnected:
			a.set(nil)
			n.SetDeviceOnline(deviceID, false)
			logrus.Info("denonavr: disconnected, reconnecting in 10s")
			select {
			case h := <-connectToHost:
				host = h
			case <-time.After(10 * time.Second):
			case <-n.Stopped():
				return
			}
		}
	}
}

// readLoop reads CR-terminated event lines from conn and updates device state.
// It returns when conn is closed or an I/O error occurs.
func readLoop(n *node.Node, conn net.Conn) {
	buf := bufio.NewReader(conn)
	for {
		line, err := buf.ReadString('\r')
		if err != nil {
			if errors.Is(err, net.ErrClosed) || errors.Is(err, io.EOF) {
				return
			}
			logrus.Errorf("denonavr: read: %s", err)
			return
		}
		line = strings.TrimRight(line, "\r\n ")
		if line == "" {
			continue
		}
		logrus.Debugf("denonavr: event: %s", line)
		if state := parseEvent(line); len(state) > 0 {
			n.UpdateState(deviceID, state)
		}
	}
}
