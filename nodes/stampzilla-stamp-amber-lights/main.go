package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"io"
	"strconv"

	log "github.com/cihub/seelog"
	"github.com/stampzilla/stampzilla-go/nodes/basenode"
	"github.com/stampzilla/stampzilla-go/protocol"
	"github.com/tarm/goserial"
	//"math/rand"
	//"time"
	"strings"
)

var node *protocol.Node
var c0 *SerialConnection

var targetColor [4]byte
var state *State = &State{[]*Device{}, make(map[string]*Sensor, 0)}

var send chan interface{} = make(chan interface{})

type SerialConnection struct {
	Name string
	Baud int
	Port io.ReadWriteCloser
}

func init() {
	// Load flags
	var dev string
	flag.StringVar(&dev, "dev", "/dev/ttyACM0", "Arduino serial port")
	flag.Parse()

	//Setup Config
	basenode.SetConfig(basenode.NewConfig())

	//Start communication with the server
	connection := basenode.Connect()
	go monitorState(node, connection)
	go serverRecv(connection)

	// Setup the serial connection
	c0 = &SerialConnection{Name: dev, Baud: 9600}
}

func main() {
	// Create new node description
	node = protocol.NewNode("stamp-amber-lights")
	node.SetState(state)
	state.Sensors["temp1"] = NewSensor("temp1", "Temperature - Bottom level", "20C")
	state.Sensors["temp2"] = NewSensor("temp2", "Temperature - Top level", "30C")
	state.Sensors["press"] = NewSensor("press", "Air pressure", "1019 hPa")

	state.AddDevice("0", "Color", []string{"color"}, "0")
	state.AddDevice("1", "Red", []string{"dim"}, "0")
	state.AddDevice("2", "Green", []string{"dim"}, "0")
	state.AddDevice("3", "Blue", []string{"dim"}, "0")

	c0.connect()

	//for {
	//select {
	//case <- time.After(time.Second):
	//state.Sensors["press"].State = strconv.FormatInt(int64(rand.Intn(40) + 980),10) +" hPa"
	//send <- node
	//}
	//}

	select {}
}

func monitorState(node *protocol.Node, connection basenode.Connection) {
	for s := range connection.State() {
		switch s {
		case basenode.ConnectionStateConnected:
			send <- node
		case basenode.ConnectionStateDisconnected:
		}
	}
}

func serverRecv(connection basenode.Connection) {

	for d := range connection.Receive() {
		processCommand(d)
	}

}

func processCommand(cmd protocol.Command) {

	type Cmd struct {
		Cmd uint16
		Arg uint32
	}

	type CmdColor struct {
		Cmd uint16
		Arg [4]byte
	}

	buf := new(bytes.Buffer)

	log.Info(cmd)

	switch cmd.Cmd {
	case "dim":
		value, _ := strconv.ParseInt(cmd.Args[1], 10, 32)

		value *= 255
		value /= 100

		switch cmd.Args[0] {
		case "1":
			targetColor[0] = byte(value)
		case "2":
			targetColor[1] = byte(value)
		case "3":
			targetColor[2] = byte(value)
		}

		err := binary.Write(buf, binary.BigEndian, &CmdColor{Cmd: 1, Arg: targetColor})
		if err != nil {
			log.Error("binary.Write failed:", err)
		}
	default:
		return
	}

	n, err := c0.Port.Write(buf.Bytes())
	if err != nil {
		log.Error(err)
	}
	log.Info("Wrote ", n, " bytes")
}

func (config *SerialConnection) connect() {

	c := &serial.Config{Name: config.Name, Baud: config.Baud}
	var err error

	config.Port, err = serial.OpenPort(c)
	if err != nil {
		log.Critical(err)
	}

	go func() {
		var incomming string = ""

		for {
			buf := make([]byte, 128)

			n, err := config.Port.Read(buf)
			if err != nil {
				log.Critical(err)
				return
			}

			incomming += string(buf[:n])

			// try to process
			for {
				n := strings.Index(incomming, ">")

				if n < 2 {
					break
				}

				msg := incomming[2:n]
				incomming = incomming[n+1:]

				pkgs := strings.Split(msg, "|")

				if len(pkgs) > 3 {
					state.Sensors["temp1"].State = pkgs[0] + " C"
					state.Sensors["temp2"].State = pkgs[3] + " C"
					state.Sensors["press"].State = pkgs[2] + " hPa"

					log.Info("IN: ", pkgs)

					send <- node
				}
			}
		}

	}()
}
