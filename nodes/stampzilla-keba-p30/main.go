package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/models/devices"
	"github.com/stampzilla/stampzilla-go/v2/pkg/node"
)

const kebaPort = "7090"

func main() {
	wg, node, _ := start()
	if node == nil {
		return
	}
	node.Wait()
	wg.Wait()
}

type sendRequest struct {
	Msg  string
	Resp chan []byte
	Err  chan error
}

func start() (*sync.WaitGroup, *node.Node, chan string) {
	listenData := make(chan []byte, 100)
	sendData := make(chan sendRequest)
	connectToIP := make(chan string)

	wg := &sync.WaitGroup{}
	node := node.New("keba-p30")

	node.OnConfig(updatedConfig(connectToIP))
	if err := node.Connect(); err != nil {
		logrus.Error(err)
		return nil, nil, nil
	}
	node.OnRequestStateChange(func(state devices.State, _ *devices.Device) error {
		var err error
		var resp []byte
		state.Float("maxCurrent", func(v float64) {
			resp, err = Send(sendData, fmt.Sprintf("currtime %.0f 1", v*1000.0))
		})
		if err != nil {
			return err
		}

		err = expectOKResponse(resp)
		if err != nil {
			return err
		}

		state.Bool("on", func(on bool) {
			if on {
				resp, err = Send(sendData, "ena 1")
			} else {
				resp, err = Send(sendData, "ena 0")
			}
		})

		if err != nil {
			return err
		}

		err = expectOKResponse(resp)
		if err != nil {
			return err
		}

		return nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	node.OnShutdown(func() {
		cancel()
	})

	syncWorker(ctx, wg, listenData, node, sendData, connectToIP)

	wg.Add(2)
	go func() {
		defer wg.Done()
		for {
			if ctx.Err() != nil {
				return
			}
			err := startListen(ctx, wg, listenData)
			if err != nil {
				logrus.Error("error start listening: ", err)
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(time.Second * 30)
		// logrus.Infof("Config OK. starting fetch loop for %s", dur)
		for {
			select {
			case <-ticker.C:
				d2, err := Send(sendData, "report 2")
				if err != nil {
					logrus.Error(err)
				}
				d3, err := Send(sendData, "report 3")
				if err != nil {
					logrus.Error(err)
				}
				err = parseAndSync(d2, d3, node)
				if err != nil {
					logrus.Error(err)
				}

			case <-node.Stopped():
				ticker.Stop()
				log.Println("Stopping keba-p30 node")
				return
			}
		}
	}()

	return wg, node, connectToIP
}

func initSender(ip string) (*net.UDPConn, error) {
	raddr, err := net.ResolveUDPAddr("udp", net.JoinHostPort(ip, kebaPort))
	if err != nil {
		return nil, err
	}

	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func Send(sendData chan sendRequest, cmd string) ([]byte, error) {
	for i := 0; i < 3; i++ {
		resp, err := send(sendData, cmd)
		if err != nil && errors.Is(err, errTimeout) {
			logrus.Error("timeout, retry in 5 sek")
			time.Sleep(100 * time.Millisecond)
			continue
		}
		return resp, err
	}
	return nil, fmt.Errorf("failed after retries")
}

func send(sendData chan sendRequest, cmd string) ([]byte, error) {
	req := sendRequest{
		Msg:  cmd,
		Err:  make(chan error),
		Resp: make(chan []byte),
	}
	sendData <- req
	select {
	case err := <-req.Err:
		return nil, err
	case r := <-req.Resp:
		return r, nil
	}

}

var errTimeout = fmt.Errorf("timeout waiting for response")

func syncWorker(ctx context.Context, wg *sync.WaitGroup, listenData chan []byte, node *node.Node, sendData chan sendRequest, connectToIP chan string) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		var conn *net.UDPConn
		responseCh := make(chan []byte)
		for {
			select {
			case <-ctx.Done():
				return
			case i := <-connectToIP:
				var err error
				conn, err = initSender(i)
				if err != nil {
					logrus.Error(err)
				}
				go func() {
					// this is the first connection after we get connectToIP
					_, err := send(sendData, "i")
					if err != nil {
						logrus.Error(err)
					}

					d2, err := send(sendData, "report 2")
					if err != nil {
						logrus.Error(err)
					}
					d3, err := send(sendData, "report 3")
					if err != nil {
						logrus.Error(err)
					}
					err = parseAndSync(d2, d3, node)
					if err != nil {
						logrus.Error(err)
					}

				}()

			case d := <-sendData:
				if conn == nil {
					logrus.Error("we are not ready to send yet")
					continue
				}
				//TODO
				// The minimum waiting time between the scheduled repetitions of any UPD
				// command is defined as follows:
				// ● t_COM_pause = 5 s

				go func() {
					tCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
					defer cancel()
					select {
					case resp := <-responseCh:
						d.Resp <- resp

					case <-tCtx.Done():
						d.Err <- errTimeout
					}
				}()
				logrus.Debugf("sending data: %s", d.Msg)
				_, err := io.Copy(conn, strings.NewReader(d.Msg))
				if err != nil {
					d.Err <- err
					continue
				}

				if d.Msg == "currtime 0 1" || d.Msg == "ena 0" {
					// The minimum waiting time after sending a disable command (e.g. ena 0) is
					// defined as follows:
					// ● t_DIS_pause = 2 s
					time.Sleep(2 * time.Second)
				} else {
					// The minimum waiting time between any two UDP commands is defined as
					// follows:
					// ● t_UDP_pause = 100 ms
					time.Sleep(100 * time.Millisecond)
				}

			case d := <-listenData:
				select {
				case responseCh <- d:
					continue
				default:
				}
				logrus.Info("no one waiting for reply for msg: ", string(d))
				//TODO
				// P30 will send status messages to the source/IP address of the last UDP
				// command it received. That means if there is only one application in the network sending commands to the charging station, the application will get the
				// information about the most important state changes without the need to poll
				// reports. P30 will provide the information about the following state changes:
				// ● “State” (see “report 2”)
				// ● “Plug” (see “report 2”)
				// ● “Input” (see “report 2”)
				// ● “Enable sys” (see “report 2”)
				// ● “Max curr” (see “report 2”)
				// ● “E pres” (see “report 3”)
				// err := parseAndSync(d, node)
				// if err != nil {
				// 	logrus.Error(err)
				// }
			}
		}
	}()
}

func parseAndSync(data2 []byte, data3 []byte, node *node.Node) error {

	rep2 := &Report2{}
	rep3 := &Report3{}

	if data2 != nil {
		err := json.Unmarshal(data2, rep2)
		if err != nil {
			return err
		}
	}

	if data3 != nil {
		err := json.Unmarshal(data3, rep3)
		if err != nil {
			return err
		}
	}

	dev := node.GetDevice("1")

	state := devices.State{}

	if data2 != nil {
		state["on"] = 0
		state["state"] = rep2.State
		state["maxCurrent"] = float64(rep2.MaxCurr) / 1000.0
		state["on"] = rep2.EnableUser == 1
	}
	if data3 != nil {
		state["L1_A"] = float64(rep3.I1) / 1000.0
		state["L2_A"] = float64(rep3.I2) / 1000.0
		state["L3_A"] = float64(rep3.I3) / 1000.0
		state["currentSessionKwh"] = float64(rep3.EPres) / 10000.0
		state["totalKwh"] = float64(rep3.ETotal) / 10000.0
	}

	if dev == nil {
		newDev := &devices.Device{
			Type: "charger",
			ID: devices.ID{
				ID: "1",
			},
			Name:   "Keba P30",
			Online: true,
			State:  state,
		}
		node.AddOrUpdate(newDev)
		return nil
	}

	logrus.Debug("sync report to device")
	node.UpdateState("1", state)
	return nil
}

func startListen(ctx context.Context, wg *sync.WaitGroup, listenData chan []byte) error {

	logrus.Infof("started udp4 listener on %s", kebaPort)
	laddr, err := net.ResolveUDPAddr("udp4", ":"+kebaPort)
	if err != nil {
		return err
	}
	l, err := net.ListenUDP("udp4", laddr)
	if err != nil {
		return err
	}
	err = l.SetReadBuffer(1024 * 2)
	if err != nil {
		return err
	}
	for {
		if ctx.Err() != nil {
			return nil
		}
		buf := make([]byte, 512)
		n, raddr, err := l.ReadFromUDP(buf)
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				return nil
			}
			logrus.Error(err)
			continue
		}
		logrus.Debugf("got UDP message from %s: %s", raddr.AddrPort().String(), string(buf[0:n]))
		listenData <- buf[0:n]
	}
}

var config = &Config{}

func updatedConfig(connectToIP chan string) node.OnFunc {
	return func(data json.RawMessage) error {
		logrus.Info("Received config from server:", string(data))

		newConf := &Config{}
		err := json.Unmarshal(data, newConf)
		if err != nil {
			return fmt.Errorf("error decoding json config: %w", err)
		}

		if newConf.IP != config.IP {
			logrus.Infof("got new IP from config %s", newConf.IP)
			connectToIP <- newConf.IP
		}
		config = newConf
		return nil
	}
}

type Config struct {
	IP string
}

func expectOKResponse(res []byte) error {
	if bytes.Equal(res, []byte("TCH-OK: done")) {
		return nil
	}
	return fmt.Errorf("expected TCH-OK, got: %s", string(res))
}
