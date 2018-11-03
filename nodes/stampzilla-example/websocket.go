package main

import (
	"context"
	"log"
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

type WebsocketClient struct {
	Conn     *websocket.Conn
	readDone chan struct{}
	//interrupt chan os.Signal
	write        chan interface{}
	wg           *sync.WaitGroup
	ctx          context.Context
	disconnected chan error
}

func NewWebsocketClient() *WebsocketClient {
	return &WebsocketClient{
		readDone:     make(chan struct{}),
		write:        make(chan interface{}),
		wg:           &sync.WaitGroup{},
		disconnected: make(chan error, 1),
	}
}

func (ws *WebsocketClient) ConnectContext(ctx context.Context, addr string) error {
	headers := http.Header{}
	headers.Add("Sec-WebSocket-Protocol", "node")

	ws.ctx = ctx

	c, _, err := websocket.DefaultDialer.DialContext(ctx, addr, headers)
	if err != nil {
		select {
		case ws.disconnected <- err:
		default:
		}
		return err
	}
	ws.Conn = c
	ws.wg.Add(2)
	go ws.readPump()
	go ws.writePump()
	return nil
}
func (ws *WebsocketClient) Wait() {
	ws.wg.Wait()
}

func (ws *WebsocketClient) Disconnected() <-chan error {
	return ws.disconnected
}

func (ws *WebsocketClient) readPump() {
	defer ws.wg.Done()
	ws.Conn.SetReadDeadline(time.Now().Add(pongWait))
	ws.Conn.SetPongHandler(func(string) error { ws.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := ws.Conn.ReadMessage()
		if err != nil {
			logrus.Error("read:", err)
			select {
			case ws.disconnected <- err:
			default:
			}

			return
		}
		logrus.Infof("recv: %s", message)
	}

}
func (wc *WebsocketClient) WriteJSON(v interface{}) {
	wc.write <- v
}

func (ws *WebsocketClient) writePump() {
	defer ws.wg.Done()
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()
	for {
		select {
		case <-ws.ctx.Done():
			return
		case t := <-ws.write:
			err := ws.Conn.WriteJSON(t)
			if err != nil {
				log.Println("write:", err)
				return
			}
		case <-ws.ctx.Done():
			err := ws.Conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				logrus.Error("write close:", err)
				return
			}
			return
		case <-ticker.C:
			if err := ws.Conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(writeWait)); err != nil {
				log.Println("ping:", err)
			}
		}
	}
}
