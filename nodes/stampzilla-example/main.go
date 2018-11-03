package main

import (
	"context"
	"flag"
	"log"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

func main() {
	flag.Parse()
	log.SetFlags(0)

	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/ws"}
	log.Printf("connecting to %s", u.String())

	client := NewWebsocketClient()
	node := NewNode(client)

	err := node.ConnectWithRetry(u.String())
	if err != nil {
		logrus.Error(err)
	}

	node.Wait()

	logrus.Info("node done...")
	logrus.Info("waiting for client to be done")
	node.Client.Wait()
	logrus.Info("client done")
}

type Node struct {
	Client *WebsocketClient
	Cancel context.CancelFunc
	wg     *sync.WaitGroup
}

func NewNode(client *WebsocketClient) *Node {
	return &Node{
		Client: client,
		wg:     &sync.WaitGroup{},
	}
}
func (n *Node) Wait() {
	n.wg.Wait()
}

func (n *Node) connect(addr string) error {
	ctx, cancel := context.WithCancel(context.Background())
	n.Cancel = cancel
	logrus.Info("Connecting to ", addr)
	err := n.Client.ConnectContext(ctx, addr)

	if err != nil {
		cancel()
		return err
	}
	return nil
}
func (n *Node) ConnectWithRetry(addr string) error {

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGQUIT, syscall.SIGTERM)

	n.wg.Add(1)
	go func() {
		defer n.wg.Done()
		for {
			select {
			case err := <-n.Client.Disconnected():
				if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
					return
				}
				logrus.Info("connection retry because: ", err)
				//TODO this makes shutdown using interrupt delayed for maximum 5 secs. Other solutions?
				time.Sleep(5 * time.Second)
				n.connect(addr)
			case <-interrupt:
				n.Cancel()
				return
			}
		}

	}()

	err := n.connect(addr)

	if err != nil {
		return err
	}

	return nil
}
