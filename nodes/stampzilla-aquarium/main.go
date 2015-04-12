package main

import (
	"flag"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/tarm/goserial"

	log "github.com/cihub/seelog"
	"github.com/stampzilla/stampzilla-go/nodes/basenode"
	"github.com/stampzilla/stampzilla-go/protocol"
)

type SerialConnection struct {
	Name string
	Baud int
	Port io.ReadWriteCloser
}

var node *protocol.Node
var state *State = &State{}
var serverConnection *basenode.Connection
var c0 *SerialConnection

func main() {
	config := basenode.NewConfig()
	basenode.SetConfig(config)

	var dev string
	flag.StringVar(&dev, "dev", "/dev/arduino", "Arduino serial port")
	flag.Parse()

	flag.Parse()

	log.Info("Starting Aquarium node")

	// Create new node description
	node = protocol.NewNode("aquarium")
	node.SetState(state)

	node.AddElement(&protocol.Element{
		Type:     protocol.ElementTypeText,
		Name:     "Water temperature",
		Feedback: `WaterTemperature`,
	})

	serverConnection = basenode.Connect()
	go monitorState(serverConnection)

	// This worker recives all incomming commands
	go serverRecv(serverConnection)

	// Setup the serial connection
	c0 = &SerialConnection{Name: dev, Baud: 9600}
	c0.connect(serverConnection)

}

// WORKER that monitors the current connection state
func monitorState(connection *basenode.Connection) {
	for s := range connection.State {
		switch s {
		case basenode.ConnectionStateConnected:
			connection.Send <- node.Node()
		case basenode.ConnectionStateDisconnected:
		}
	}
}

// WORKER that recives all incomming commands
func serverRecv(connection *basenode.Connection) {
	for d := range connection.Receive {
		if err := processCommand(d); err != nil {
			log.Error(err)
		}
	}
}

func processCommand(cmd protocol.Command) error {

	return nil
}

func (config *SerialConnection) connect(connection *basenode.Connection) {

	c := &serial.Config{Name: config.Name, Baud: config.Baud}
	var err error

	config.Port, err = serial.OpenPort(c)
	if err != nil {
		log.Critical(err)
		return
	}

	go func() {
		var incomming string = ""
		var updateInhibit bool = false
		var changed bool = false

		for {
			buf := make([]byte, 128)

			n, err := config.Port.Read(buf)
			if err != nil {
				log.Critical(err)
				return
			}

			incomming += string(buf[:n])

			for {
				n := strings.Index(incomming, "\r")

				if n < 0 {
					break
				}

				msg := strings.TrimSpace(incomming[:n])
				incomming = incomming[n+1:]
				//log.Info(msg)

				values := strings.Split(msg, "|")
				if len(values) != 3 {
					continue
				}

				value, err := strconv.ParseFloat(values[0], 32)
				if err != nil {
					continue
				}
				if float32(value) != state.WaterTemperature {
					changed = true
				}
				state.WaterTemperature = float32(value)

				value, err = strconv.ParseFloat(values[1], 32)
				if err != nil {
					continue
				}
				if state.Filling != (value > 0) {
					changed = true

				}
				state.Filling = value > 0

				value, err = strconv.ParseFloat(values[2], 32)
				if err != nil {
					continue
				}
				if state.Cooling != float32(value) {
					changed = true
				}
				state.Cooling = float32(value)

				if !updateInhibit && changed {
					changed = false
					connection.Send <- node.Node()
					updateInhibit = true
					//log.Warn("Invalid pacakge: ", msg)
					go func() {
						<-time.After(time.Millisecond * 100)
						updateInhibit = false
					}()
				}

			}
		}
	}()

	select {}
}
