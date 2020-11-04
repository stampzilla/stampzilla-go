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
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-spc/edp"
	"github.com/stampzilla/stampzilla-go/pkg/node"
)

func main() {
	node := node.New("spc")
	connectToPort := make(chan string)
	node.OnConfig(updatedConfig(connectToPort))
	err := node.Connect()
	if err != nil {
		logrus.Error(err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	wg := startListen(ctx, node, connectToPort)

	node.OnShutdown(func() {
		cancel()
	})

	node.Wait()
	wg.Wait()
}

func startListen(ctx context.Context, node *node.Node, connectToPort chan string) *sync.WaitGroup {
	wg := &sync.WaitGroup{}

	listen := func(port string) net.PacketConn {
		logrus.Infof("started udp4 server on %s", port)
		l, err := net.ListenPacket("udp4", ":"+port)
		if err != nil {
			logrus.Error(err)
			return nil
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				buf := make([]byte, 1024)
				n, _, err := l.ReadFrom(buf)
				if err != nil {
					if strings.Contains(err.Error(), "use of closed network connection") {
						return
					}
					logrus.Error(err)
					continue
				}
				logrus.Debug("string", string(buf[0:n]))
				if logrus.GetLevel() >= logrus.DebugLevel {
					fmt.Println(hex.Dump(buf[0:n]))
				}
				pkg, err := edp.Decode(buf[0:n])
				if err != nil {
					logrus.Error(err)
					continue
				}

				dev := node.GetDevice(pkg.ID)
				newDev := edp.GenerateDevice(pkg)

				if dev == nil && newDev != nil {
					node.AddOrUpdate(newDev)
					continue
				}

				if newDev == nil {
					logrus.Warnf("unsupported packet class %s data: %s", pkg.Class, string(buf[23:]))
					continue
				}

				node.UpdateState(pkg.ID, newDev.State)
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

	return wg
}

var config = &Config{}

func updatedConfig(connectToPort chan string) node.OnFunc {
	return func(data json.RawMessage) error {
		logrus.Info("Received config from server:", string(data))

		newConf := &Config{}
		err := json.Unmarshal(data, newConf)
		if err != nil {
			return err
		}

		if newConf.EDPPort != config.EDPPort {
			fmt.Println("ip changed. lets connect to that instead")
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
