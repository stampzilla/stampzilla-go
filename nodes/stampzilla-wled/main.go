package main

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/models/devices"
	"github.com/stampzilla/stampzilla-go/v2/pkg/node"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

const deviceID = "1"

var errNotConnected = errors.New("wled: not connected")

func main() {
	n, _ := start()
	if n == nil {
		return
	}
	n.Wait()
}

// wledConn holds the live WebSocket connection and serialises writes.
// nhooyr.io/websocket allows one concurrent reader and one concurrent writer;
// the mutex ensures writes from OnRequestStateChange and the keepalive goroutine
// do not overlap.
type wledConn struct {
	mu   sync.Mutex
	conn *websocket.Conn
}

// set replaces the active connection, closing any previous one.
// Pass nil to clear without opening a new connection.
func (w *wledConn) set(conn *websocket.Conn) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.conn != nil {
		w.conn.Close(websocket.StatusNormalClosure, "replacing connection")
	}
	w.conn = conn
}

// send JSON-marshals payload and writes it to the WebSocket.
func (w *wledConn) send(ctx context.Context, payload map[string]any) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.conn == nil {
		return errNotConnected
	}
	wctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return wsjson.Write(wctx, w.conn, payload)
}

// ping sends a WebSocket ping to detect half-open connections.
func (w *wledConn) ping(ctx context.Context) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.conn == nil {
		return errNotConnected
	}
	pctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return w.conn.Ping(pctx)
}

var nodeConfig = &Config{}

func updatedConfig(connectToHost chan<- string) node.OnFunc {
	return func(data json.RawMessage) error {
		newConf := &Config{}
		if err := json.Unmarshal(data, newConf); err != nil {
			return err
		}
		if newConf.Host != nodeConfig.Host {
			logrus.Infof("wled: new host from config: %s", newConf.Host)
			nodeConfig = newConf
			connectToHost <- newConf.Host
		}
		return nil
	}
}

func start() (*node.Node, *wledConn) {
	connectToHost := make(chan string, 1)
	w := &wledConn{}

	n := node.New("wled")
	n.OnConfig(updatedConfig(connectToHost))

	if err := n.Connect(); err != nil {
		logrus.Error(err)
		return nil, nil
	}

	// baseCtx is cancelled when the node shuts down.
	baseCtx, baseCancel := context.WithCancel(context.Background())
	go func() {
		<-n.Stopped()
		baseCancel()
	}()

	n.OnRequestStateChange(func(state devices.State, _ *devices.Device) error {
		payload := map[string]any{}
		state.Bool("on", func(on bool) {
			payload["on"] = on
		})
		state.Float("brightness", func(b float64) {
			payload["bri"] = briFromFloat(b)
		})
		if len(payload) == 0 {
			return nil
		}
		return w.send(baseCtx, payload)
	})

	ensureDevice(n)
	go connectionWorker(baseCtx, n, w, connectToHost)

	return n, w
}

// ensureDevice registers the WLED device with the node if it does not exist yet.
func ensureDevice(n *node.Node) {
	if n.GetDevice(deviceID) != nil {
		return
	}
	n.AddOrUpdate(&devices.Device{
		Type:   "light",
		ID:     devices.ID{ID: deviceID},
		Name:   "WLED",
		Online: false,
		Traits: []string{"OnOff", "Brightness"},
		State:  devices.State{"on": false, "brightness": 0.0},
	})
}

// connectionWorker maintains a persistent WebSocket connection to the WLED device.
// It reconnects on error and responds to host changes via connectToHost.
func connectionWorker(ctx context.Context, n *node.Node, w *wledConn, connectToHost <-chan string) {
	var host string

	// Wait for the first host before attempting any connection.
	select {
	case h := <-connectToHost:
		host = h
	case <-ctx.Done():
		return
	}

	for {
		if ctx.Err() != nil {
			return
		}

		addr := wsURL(host)
		logrus.Infof("wled: connecting to %s", addr)

		dialCtx, dialCancel := context.WithTimeout(ctx, 10*time.Second)
		conn, _, err := websocket.Dial(dialCtx, addr, nil)
		dialCancel()

		if err != nil {
			logrus.Errorf("wled: dial %s: %s", addr, err)
			n.SetDeviceOnline(deviceID, false)
			select {
			case h := <-connectToHost:
				host = h
			case <-time.After(10 * time.Second):
			case <-ctx.Done():
				return
			}
			continue
		}

		logrus.Infof("wled: connected to %s", addr)
		w.set(conn)
		ensureDevice(n)
		n.SetDeviceOnline(deviceID, true)

		// Request the current full state so we populate the device immediately.
		if err := w.send(ctx, map[string]any{"v": true}); err != nil {
			logrus.Warnf("wled: initial state request failed: %s", err)
		}

		// Per-connection context: cancellation stops the readLoop and keepalive.
		connCtx, connCancel := context.WithCancel(ctx)

		// Keepalive: ping every 30s to detect half-open TCP connections.
		go func() {
			ticker := time.NewTicker(30 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					if err := w.ping(connCtx); err != nil {
						logrus.Warnf("wled: ping failed: %s — reconnecting", err)
						connCancel()
						return
					}
				case <-connCtx.Done():
					return
				}
			}
		}()

		// Read loop: blocks until the connection closes or connCtx is cancelled.
		disconnected := make(chan struct{})
		go func() {
			defer close(disconnected)
			readLoop(connCtx, n, conn)
		}()

		select {
		case <-ctx.Done():
			connCancel()
			w.set(nil)
			<-disconnected
			return

		case h := <-connectToHost:
			host = h
			connCancel()
			<-disconnected
			w.set(nil)
			// continue to reconnect with the new host

		case <-disconnected:
			connCancel()
			w.set(nil)
			n.SetDeviceOnline(deviceID, false)
			logrus.Info("wled: disconnected, reconnecting in 10s")
			select {
			case h := <-connectToHost:
				host = h
			case <-time.After(10 * time.Second):
			case <-ctx.Done():
				return
			}
		}
	}
}

// readLoop reads JSON messages pushed by WLED and updates device state.
// It returns when conn is closed or ctx is cancelled.
func readLoop(ctx context.Context, n *node.Node, conn *websocket.Conn) {
	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			if ctx.Err() == nil {
				logrus.Errorf("wled: read: %s", err)
			}
			return
		}
		state, err := parseWLEDState(data)
		if err != nil {
			logrus.Warnf("wled: parse message: %s", err)
			continue
		}
		logrus.Debugf("wled: state update: %v", state)
		n.UpdateState(deviceID, state)
	}
}
