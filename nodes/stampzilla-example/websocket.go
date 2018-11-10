package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/models"
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

type writeData struct {
	data interface{}
	err  chan error
}

type WebsocketClient struct {
	conn            *websocket.Conn
	TLSClientConfig *tls.Config
	readDone        chan struct{}
	//interrupt chan os.Signal
	write        chan writeData
	read         chan *models.Message
	wg           *sync.WaitGroup
	disconnected chan error
}

func NewWebsocketClient() *WebsocketClient {
	return &WebsocketClient{
		readDone:     make(chan struct{}),
		write:        make(chan writeData),
		read:         make(chan *models.Message, 100),
		wg:           &sync.WaitGroup{},
		disconnected: make(chan error),
	}
}

func (ws *WebsocketClient) ConnectContext(ctx context.Context, addr string, headers http.Header) error {
	var err error
	var c *websocket.Conn
	if ws.TLSClientConfig != nil {
		dialer := &websocket.Dialer{
			Proxy:            http.ProxyFromEnvironment,
			HandshakeTimeout: 45 * time.Second,
			TLSClientConfig:  ws.TLSClientConfig,
		}
		c, _, err = dialer.DialContext(ctx, addr, headers)
	} else {
		c, _, err = websocket.DefaultDialer.DialContext(ctx, addr, headers)
	}
	if err != nil {
		select {
		case ws.disconnected <- err:
		default:
		}
		return err
	}
	ws.conn = c
	ws.wg.Add(2)
	go ws.readPump()
	go ws.writePump(ctx)
	return nil
}
func (ws *WebsocketClient) Wait() {
	ws.wg.Wait()
}

func (ws *WebsocketClient) Message() <-chan *models.Message {
	return ws.read
}

// WaitForMessage is a helper method to wait for a specific message type
func (ws *WebsocketClient) WaitForMessage(msgType string, dst interface{}) error {

	for msg := range ws.Message() {
		if msg.Type == msgType {
			return json.Unmarshal(msg.Body, dst)
		}
	}
	return nil
}
func (ws *WebsocketClient) Disconnected() <-chan error {
	return ws.disconnected
}

func (ws *WebsocketClient) readPump() {
	defer ws.wg.Done()
	ws.conn.SetReadDeadline(time.Now().Add(pongWait))
	ws.conn.SetPongHandler(func(string) error { ws.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := ws.conn.ReadMessage()
		if err != nil {
			logrus.Error("read:", err)
			select {
			case ws.disconnected <- err:
			default:
			}

			return
		}
		logrus.Infof("recv: %s", message)
		msg, err := models.ParseMessage(message)
		if err != nil {
			logrus.Error("ParseMessage error: ", err)
			continue
		}
		select {
		case ws.read <- msg:
		default:
		}
	}

}
func (wc *WebsocketClient) WriteJSON(v interface{}) error {
	errCh := make(chan error, 1)
	select {
	case wc.write <- writeData{data: v, err: errCh}:
	default:
		errCh <- fmt.Errorf("no one listening on write channel")
	}
	return <-errCh
}

func (ws *WebsocketClient) writePump(ctx context.Context) {
	defer ws.wg.Done()
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()
	for {
		select {
		case t := <-ws.write:
			err := ws.conn.WriteJSON(t.data)
			t.err <- err
		case <-ctx.Done():
			err := ws.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				logrus.Error("write close:", err)
				return
			}
			return
		case <-ticker.C:
			if err := ws.conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(writeWait)); err != nil {
				logrus.Error("ping:", err)
			}
		}
	}
}
