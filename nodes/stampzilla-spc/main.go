package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-spc/edp"
	"github.com/stampzilla/stampzilla-go/v2/pkg/node"
)

func main() {
	wg, node, _ := start()
	if node == nil {
		return
	}
	node.Wait()
	wg.Wait()
}

func start() (*sync.WaitGroup, *node.Node, chan string) {
	node := node.New("spc")
	connectToPort := make(chan string)
	node.OnConfig(updatedConfig(connectToPort))
	if err := node.Connect(); err != nil {
		logrus.Error(err)
		return nil, nil, nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	data := make(chan []byte, 100)
	wg := &sync.WaitGroup{}
	syncWorker(ctx, wg, data, node)
	startListen(ctx, wg, connectToPort, data)

	node.OnShutdown(func() {
		cancel()
	})

	return wg, node, connectToPort
}

func syncWorker(ctx context.Context, wg *sync.WaitGroup, data chan []byte, node *node.Node) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case d := <-data:
				err := decodeAndSync(d, node)
				if err != nil {
					logrus.Error(err)
					return
				}
			}
		}
	}()
}

func decodeAndSync(buf []byte, node *node.Node) error {
	logrus.Debug("string", string(buf))
	if logrus.GetLevel() >= logrus.DebugLevel {
		fmt.Println(hex.Dump(buf)) // nolint
	}
	pkg, err := edp.Decode(buf)
	if err != nil {
		return fmt.Errorf("error decodingn pkg: %w", err)
	}

	dev := node.GetDevice(pkg.ID)
	newDev := edp.GenerateDevice(pkg)

	if dev == nil && newDev != nil {
		node.AddOrUpdate(newDev)
		return nil
	}

	if newDev == nil {
		logrus.Warnf("unsupported packet class %s data: %s", pkg.Class, string(buf[23:]))
		return nil
	}
	node.UpdateState(pkg.ID, newDev.State)
	return nil
}

func startListen(ctx context.Context, wg *sync.WaitGroup, connectToPort chan string, data chan []byte) {
	listen := func(port string) net.PacketConn {
		logrus.Infof("started udp4 server on %s", port)
		l, err := net.ListenPacket("udp4", ":"+port)
		if err != nil {
			logrus.Error(err)
			return nil
		}
		if c, ok := l.(*net.UDPConn); ok {
			rb := 1024 * 1024
			logrus.Infof("setting ReadBuffer on udp conn to: %d", rb)
			c.SetReadBuffer(rb)
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				buf := make([]byte, 1440)
				n, _, err := l.ReadFrom(buf)
				if err != nil {
					if strings.Contains(err.Error(), "use of closed network connection") {
						return
					}
					logrus.Error(err)
					continue
				}
				data <- buf[0:n]
			}
		}()
		return l
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		var l net.PacketConn
		defer func() {
			if l != nil {
				l.Close()
			}
		}()
		for {
			select {
			case port := <-connectToPort:
				if l != nil {
					l.Close()
				}
				l = listen(port)
			case <-ctx.Done():
				return
			}
		}
	}()
}

var config = &Config{}

func updatedConfig(connectToPort chan string) node.OnFunc {
	return func(data json.RawMessage) error {
		logrus.Info("Received config from server:", string(data))

		newConf := &Config{}
		err := json.Unmarshal(data, newConf)
		if err != nil {
			return fmt.Errorf("error decoding json config: %w", err)
		}

		if newConf.EDPPort != config.EDPPort {
			logrus.Infof("got new EDPPort from config %s", newConf.EDPPort)
			connectToPort <- newConf.EDPPort
		}
		config = newConf
		return nil
	}
}

type Config struct {
	EDPPort string
}

/*
Config to put into gui:
{
	EDPPort: "1234"
}

*/
