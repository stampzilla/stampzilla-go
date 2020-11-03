package main

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models/devices"
	"github.com/stampzilla/stampzilla-go/pkg/node"
)

func main() {
	node := node.New("spc")
	node.OnConfig(updatedConfig)
	err := node.Connect()
	if err != nil {
		logrus.Error(err)
		return
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)

	l, err := net.ListenPacket("udp4", ":8080")
	if err != nil {
		fmt.Println(err)
		logrus.Error(err)
		return
	}

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
			// logrus.Infof("read %d from %s\n", n, addr)
			// logrus.Info("string", string(buf[0:n]))
			// fmt.Println(hex.Dump(buf[0:n]))
			// fmt.Println(printHex(buf[0:n]))
			edp, err := decode(buf[0:n])
			if err != nil {
				logrus.Error(err)
				continue
			}

			dev := node.GetDevice(edp.ID)
			state := devices.State{}

			switch edp.Action {
			case "open":
				// TODO call it on? or open? or triggered? or active? :)
				state["on"] = true
			case "close":
				state["on"] = false
			}

			if dev == nil {
				node.AddOrUpdate(&devices.Device{
					Type: "sensor",
					ID: devices.ID{
						ID: edp.ID,
					},
					Name:   edp.Name,
					Online: true,
					State:  state,
				})

				continue
			}

			node.UpdateState(edp.ID, state)
		}
	}()

	node.OnShutdown(func() {
		l.Close()
	})

	node.Wait()
	wg.Wait()
}

// func listenUDP(ctx context.Context) error {

//}

var config = &Config{}

func updatedConfig(data json.RawMessage) error {
	logrus.Info("Received config from server:", string(data))

	newConf := &Config{}
	err := json.Unmarshal(data, newConf)
	if err != nil {
		return err
	}

	// example when we change "global" config
	if newConf.EDPPort != config.EDPPort {
		fmt.Println("ip changed. lets connect to that instead")
		// TODO stop udp server and start it on new port
	}

	config = newConf
	logrus.Info("Config is now: ", config)

	return nil
}

type Config struct {
	EDPPort string
}

/*
Config to put into gui:
{
	EDPPort: "5432"
}

*/
