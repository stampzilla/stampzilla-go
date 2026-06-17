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
)

func main() {
	n, mgr := start()
	if n == nil {
		return
	}
	n.Wait()
	mgr.close()
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

// DeviceConfig represents configuration for a single AVR device.
type DeviceConfig struct {
	Host string `json:"host"`
	Name string `json:"name,omitempty"`
}

// Config is the per-node application config pushed from the stampzilla server.
type Config struct {
	Devices map[string]DeviceConfig `json:"devices"`
}

type deviceWorker struct {
	id     string
	name   string
	host   string
	cancel context.CancelFunc
	avr    *avr
}

type manager struct {
	mu      sync.Mutex
	node    *node.Node
	workers map[string]*deviceWorker
}

func newManager(n *node.Node) *manager {
	return &manager{
		node:    n,
		workers: make(map[string]*deviceWorker),
	}
}

func (m *manager) getAvr(id string) *avr {
	m.mu.Lock()
	defer m.mu.Unlock()
	if w, ok := m.workers[id]; ok {
		return w.avr
	}
	return nil
}

func (m *manager) close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, w := range m.workers {
		w.cancel()
		w.avr.close()
	}
}

func (m *manager) updatedConfig(data json.RawMessage) error {
	newConf := &Config{}
	if err := json.Unmarshal(data, newConf); err != nil {
		return fmt.Errorf("denonavr: bad config: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Stop workers that are not in the new config or have changed hosts.
	for id, w := range m.workers {
		newDev, ok := newConf.Devices[id]
		if !ok || newDev.Host != w.host {
			logrus.Infof("denonavr [%s]: stopping worker (%s)", id, w.host)
			w.cancel()
			w.avr.close()
			setDeviceOnline(m.node, id, false)
			delete(m.workers, id)
		}
	}

	// Start or update workers.
	for id, dev := range newConf.Devices {
		if w, ok := m.workers[id]; !ok {
			logrus.Infof("denonavr [%s]: starting worker (%s - %s)", id, dev.Name, dev.Host)
			ensureDevice(m.node, id, dev.Name)

			ctx, cancel := context.WithCancel(context.Background())
			a := &avr{}
			m.workers[id] = &deviceWorker{
				id:     id,
				name:   dev.Name,
				host:   dev.Host,
				cancel: cancel,
				avr:    a,
			}

			go connectionWorker(ctx, m.node, id, dev.Host, a)
		} else if dev.Name != w.name {
			// Update name without restarting the connection
			logrus.Infof("denonavr [%s]: updating device name (%s -> %s)", id, w.name, dev.Name)
			w.name = dev.Name
			ensureDevice(m.node, id, dev.Name)
		}
	}

	return nil
}

func start() (*node.Node, *manager) {
	n := node.New("denonavr")
	mgr := newManager(n)

	n.OnConfig(mgr.updatedConfig)

	if err := n.Connect(); err != nil {
		logrus.Error(err)
		return nil, nil
	}

	n.OnRequestStateChange(func(state devices.State, d *devices.Device) error {
		a := mgr.getAvr(d.ID.ID)
		if a == nil {
			return fmt.Errorf("denonavr: device %s not found", d.ID.ID)
		}

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

	return n, mgr
}

// ensureDevice registers the Denon AVR device with the node if it does not exist yet.
var ensureDevice = func(n *node.Node, id string, name string) {
	if name == "" {
		name = "Denon AVR"
	}
	if dev := n.GetDevice(id); dev != nil {
		if dev.Name != name {
			dev.Name = name
			n.AddOrUpdate(dev)
		}
		return
	}
	n.AddOrUpdate(&devices.Device{
		Type:   "mediaplayer",
		ID:     devices.ID{ID: id},
		Name:   name,
		Online: false,
		Traits: []string{"OnOff", "Volume"},
		State:  devices.State{"on": false, "volume": 0.0, "source": ""},
	})
}

var setDeviceOnline = func(n *node.Node, id string, online bool) {
	n.SetDeviceOnline(id, online)
}

// normaliseHost appends the default telnet port (23) when none is present.
func normaliseHost(host string) string {
	if _, _, err := net.SplitHostPort(host); err != nil {
		return net.JoinHostPort(host, defaultPort)
	}
	return host
}

// connectionWorker maintains a persistent TCP connection to the Denon receiver.
// It reconnects automatically on network errors until its context is canceled.
func connectionWorker(ctx context.Context, n *node.Node, id string, host string, a *avr) {
	for {
		if strings.TrimSpace(host) == "" {
			setDeviceOnline(n, id, false)
			return
		}

		addr := normaliseHost(host)
		logrus.Infof("denonavr [%s]: connecting to %s", id, addr)

		dialCtx, dialCancel := context.WithTimeout(ctx, 10*time.Second)
		conn, err := (&net.Dialer{}).DialContext(dialCtx, "tcp", addr)
		dialCancel()

		if err != nil {
			logrus.Errorf("denonavr [%s]: dial %s: %s", id, addr, err)
			setDeviceOnline(n, id, false)
			select {
			case <-time.After(10 * time.Second):
			case <-ctx.Done():
				return
			}
			continue
		}

		logrus.Infof("denonavr [%s]: connected to %s", id, addr)
		a.set(conn)
		setDeviceOnline(n, id, true)

		// Query initial state.
		_ = a.send("PW?")
		_ = a.send("MV?")
		_ = a.send("SI?")

		// Start an async reader that feeds incoming events into device state.
		disconnected := make(chan struct{})
		go func() {
			defer close(disconnected)
			readLoop(n, id, conn)
		}()

		// Block until the link drops, or we shut down.
		select {
		case <-ctx.Done():
			a.close() // triggers readLoop to exit with an I/O error
			<-disconnected
			return

		case <-disconnected:
			a.set(nil)
			setDeviceOnline(n, id, false)
			logrus.Infof("denonavr [%s]: disconnected, reconnecting in 10s", id)
			select {
			case <-time.After(10 * time.Second):
			case <-ctx.Done():
				return
			}
		}
	}
}

// readLoop reads CR-terminated event lines from conn and updates device state.
// It returns when conn is closed or an I/O error occurs.
func readLoop(n *node.Node, id string, conn net.Conn) {
	buf := bufio.NewReader(conn)
	for {
		line, err := buf.ReadString('\r')
		if err != nil {
			if errors.Is(err, net.ErrClosed) || errors.Is(err, io.EOF) {
				return
			}
			logrus.Errorf("denonavr [%s]: read: %s", id, err)
			return
		}
		line = strings.TrimRight(line, "\r\n ")
		if line == "" {
			continue
		}
		logrus.Debugf("denonavr [%s]: event: %s", id, line)
		if state := parseEvent(line); len(state) > 0 {
			n.UpdateState(id, state)
		}
	}
}
