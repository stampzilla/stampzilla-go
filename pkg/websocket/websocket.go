package websocket

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 10 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	reconnectWait = 2 * time.Second
)

// Websocket implements a websocket client
type Websocket interface {
	OnConnect(cb func())
	ConnectContext(ctx context.Context, addr string, headers http.Header) error
	ConnectWithRetry(parentCtx context.Context, addr string, headers http.Header)
	Wait()
	Read() <-chan []byte
	// WriteJSON writes interface{} encoded as JSON to our connection
	WriteJSON(v interface{}) error
	SetTLSConfig(c *tls.Config)
}

type websocketClient struct {
	conn            *websocket.Conn
	tlsClientConfig *tls.Config
	write           chan func()
	read            chan []byte
	wg              *sync.WaitGroup
	disconnected    chan error
	connected       chan struct{}
	onConnect       func()
}

// New creates a new Websocket
func New() Websocket {
	return &websocketClient{
		write:        make(chan func()),
		read:         make(chan []byte, 100),
		wg:           &sync.WaitGroup{},
		disconnected: make(chan error),
		connected:    make(chan struct{}),
	}
}

func (ws *websocketClient) SetTLSConfig(c *tls.Config) {
	ws.tlsClientConfig = c
}

func (ws *websocketClient) OnConnect(cb func()) {
	ws.onConnect = cb
}

func (ws *websocketClient) ConnectContext(ctx context.Context, addr string, headers http.Header) error {
	var err error
	var c *websocket.Conn
	logrus.Info("websocket: connecting to ", addr)
	if ws.tlsClientConfig != nil {
		dialer := &websocket.Dialer{
			Proxy:            http.ProxyFromEnvironment,
			HandshakeTimeout: 45 * time.Second,
			TLSClientConfig:  ws.tlsClientConfig,
		}
		c, _, err = dialer.DialContext(ctx, addr, headers)
	} else {
		c, _, err = websocket.DefaultDialer.DialContext(ctx, addr, headers)
	}
	if err != nil {
		ws.wasDisconnected(err)
		return err
	}
	logrus.Infof("websocket: connected to %s", addr)
	ws.wasConnected()
	ws.conn = c
	ws.readPump()
	ws.writePump(ctx) <- struct{}{}

	if ws.onConnect != nil {
		ws.onConnect()
	}
	return nil
}

// ConnectWithRetry tries to connect and blocks until connected.
// if disconnected because an error tries to reconnect again every 5th second
func (ws *websocketClient) ConnectWithRetry(parentCtx context.Context, addr string, headers http.Header) {

	ctx, cancel := context.WithCancel(parentCtx)
	ws.wg.Add(1)
	go func() {
		defer ws.wg.Done()
		for {
			select {
			case <-parentCtx.Done():
				logrus.Info("websocket: stopping reconnect because err: ", parentCtx.Err())
				return
			case err := <-ws.disconnected:
				cancel() // Stop any write/read pumps so we dont get duplicate write panic
				logrus.Error("websocket: disconnected")
				if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
					logrus.Info("websocket: Skipping reconnect due to CloseNormalClosure")
					return
				}
				logrus.Info("websocket: Reconnect because error: ", err)
				go func() {
					time.Sleep(5 * time.Second)
					ctx, cancel = context.WithCancel(parentCtx)
					err := ws.ConnectContext(ctx, addr, headers)
					if err != nil {
						logrus.Error("websocket: Reconnect failed with error: ", err)
					}
				}()
			}
		}
	}()
	go ws.ConnectContext(ctx, addr, headers)
	<-ws.connected
	return
}

func (ws *websocketClient) Wait() {
	ws.wg.Wait()
}

func (ws *websocketClient) Read() <-chan []byte {
	return ws.read
}

// WriteJSON writes interface{} encoded as JSON to our connection
func (ws *websocketClient) WriteJSON(v interface{}) error {
	errCh := make(chan error, 1)
	select {
	case ws.write <- func() {
		err := ws.conn.WriteJSON(v)
		errCh <- err
	}:
	default:
		errCh <- fmt.Errorf("websocket: no one listening on write channel")
	}
	return <-errCh
}

func (ws *websocketClient) readPump() {
	go func() {
		ws.wg.Add(1)
		defer ws.wg.Done()
		ws.conn.SetReadDeadline(time.Now().Add(pongWait))
		ws.conn.SetPongHandler(func(string) error { ws.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
		for {
			_, message, err := ws.conn.ReadMessage()
			if err != nil {
				logrus.Error("websocket: readPump error:", err)
				ws.wasDisconnected(err)
				return
			}
			logrus.Debugf("websocket: readPump got msg: %s", message)
			select {
			case ws.read <- message:
			default:
			}
		}
	}()
}

func (ws *websocketClient) writePump(ctx context.Context) chan struct{} {
	ready := make(chan struct{})
	go func() {
		ws.wg.Add(1)
		defer ws.wg.Done()
		ticker := time.NewTicker(pingPeriod)
		defer ticker.Stop()
		for {
			select {
			case t := <-ws.write:
				t()
			case <-ctx.Done():
				logrus.Error("websocket: Stopping writePump because err: ", ctx.Err())
				err := ws.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				if err != nil {
					logrus.Error("websocket: write close:", err)
					return
				}
				return
			case <-ticker.C:
				if err := ws.conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(writeWait)); err != nil {
					logrus.Error("websocket: ping:", err)
				}
			case <-ready:
			}
		}
	}()
	return ready
}

func (ws *websocketClient) wasDisconnected(err error) {
	select {
	case ws.disconnected <- err:
	default:
	}
}

func (ws *websocketClient) wasConnected() {
	select {
	case ws.connected <- struct{}{}:
	default:
	}
}
