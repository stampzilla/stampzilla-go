package main

import (
	"flag"
	"fmt"
	"io"
	"reflect"
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
var ard *SerialConnection
var disp *SerialConnection
var frameCount int

func main() {
	config := basenode.NewConfig()
	basenode.SetConfig(config)

	var arddev string
	var dispdev string
	flag.StringVar(&arddev, "arduino-dev", "/dev/arduino", "Arduino serial port")
	flag.StringVar(&dispdev, "display-dev", "/dev/poleDisplay", "Display serial port")

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
	node.AddElement(&protocol.Element{
		Type:     protocol.ElementTypeText,
		Name:     "Water level - OK",
		Feedback: `WaterLevelOk`,
	})
	node.AddElement(&protocol.Element{
		Type:     protocol.ElementTypeText,
		Name:     "Filter - OK",
		Feedback: `FilterOk`,
	})

	node.AddElement(&protocol.Element{
		Type:     protocol.ElementTypeText,
		Name:     "Cooling",
		Feedback: `Cooling`,
	})
	node.AddElement(&protocol.Element{
		Type: protocol.ElementTypeToggle,
		Name: "Heating",
		Command: &protocol.Command{
			Cmd:  "Heating",
			Args: []string{},
		},
		Feedback: `Heating`,
	})

	node.AddElement(&protocol.Element{
		Type: protocol.ElementTypeToggle,
		Name: "Skimmer",
		Command: &protocol.Command{
			Cmd:  "Skimmer",
			Args: []string{},
		},
		Feedback: `Skimmer`,
	})
	node.AddElement(&protocol.Element{
		Type: protocol.ElementTypeToggle,
		Name: "Circulation pumps",
		Command: &protocol.Command{
			Cmd:  "CirculationPumps",
			Args: []string{},
		},
		Feedback: `CirculationPumps`,
	})

	serverConnection = basenode.Connect()
	go monitorState(serverConnection)

	// This worker recives all incomming commands
	go serverRecv(serverConnection)

	// Setup the serial connection
	ard = &SerialConnection{Name: arddev, Baud: 115200}
	ard.run(serverConnection, func(data string, connection *basenode.Connection) {
		processArduinoData(data, connection)
	})

	disp = &SerialConnection{Name: dispdev, Baud: 9600}
	disp.run(serverConnection, func(data string, connection *basenode.Connection) {
		log.Debug("Incomming from display", data)
	})

	go updateDisplay(serverConnection, disp)

	go func() {
		for {
			fmt.Print("\r", frameCount)
			<-time.After(time.Millisecond * 10)
		}
	}()

	select {}
}

var updateInhibit bool = false
var changed bool = false

func processArduinoData(msg string, connection *basenode.Connection) { // {{{
	//log.Debug(msg)
	frameCount++

	var prevState State = *state

	values := strings.Split(msg, "|")
	if len(values) != 6 {
		return
	}

	// Temperature// {{{
	value, err := strconv.ParseFloat(values[0], 32)
	if err != nil {
		return
	}
	if value != state.WaterTemperature {
		changed = true
	}
	state.WaterTemperature = value // }}}

	// Filling stat0e // {{{
	value, err = strconv.ParseFloat(values[1], 32)
	if err != nil {
		return
	}
	if state.FillingTime != value {
		changed = true
	}
	state.FillingTime = value // }}}

	// Cooling %// {{{
	value, err = strconv.ParseFloat(values[2], 32)
	if err != nil {
		return
	}
	if state.Cooling != value {
		changed = true
	}
	state.Cooling = value // }}}

	bits := values[3][0]

	state.CirculationPumps = bits&0x01 != 0
	state.Skimmer = bits&0x02 != 0
	state.Heating = bits&0x04 != 0
	state.Filling = bits&0x08 == 0
	state.WaterLevelOk = bits&0x0F == 0
	state.FilterOk = bits&0x10 == 0

	// Check if something have changed
	if !reflect.DeepEqual(prevState, *state) {
		changed = true
	}

	if !updateInhibit && changed {
		changed = false
		//fmt.Print("\n")

		go func() {
			select {
			case connection.Send <- node.Node():
			case <-time.After(time.Second * 1):
				log.Warn("TIMEOUT: Failed to send update to server")
			}
		}()

		updateInhibit = true
		//log.Warn("Invalid pacakge: ", msg)
		go func() {
			<-time.After(time.Millisecond * 200)
			updateInhibit = false
		}()
	}
}                                                                               // }}}
func updateDisplay(connection *basenode.Connection, serial *SerialConnection) { // {{{
	for {
		select {
		case <-time.After(time.Second / 15):
			if serial.Port == nil {
				continue
			}

			var msg, row1, row2 string
			var wLevel string = "OK"

			//if !state.WaterLevelOk {
			//wLevel = "WLVL"
			//}
			if state.FillingTime > 0 {
				wLevel = "FILL ERR"
				if state.FillingTime < 20000 {
					wLevel = "FILLING"
				}
			}

			//msg += "\x1B" // Reset
			//msg += "\x0E" // Clear
			//msg += "\x13" // Cursor on
			msg += "\x11" // Character over write mode
			//msg += "\x12" // Vertical scroll mode
			//msg += "\x13" // Cursor on
			msg += "\x14" // Cursor off

			msg += "\r"

			row1 = time.Now().Format("15:04:05") + strings.Repeat(" ", 12-len(wLevel)) + wLevel
			row2 = "Temp " + strconv.FormatFloat(state.WaterTemperature, 'f', 2, 64) + " Cool " + strconv.FormatFloat(state.Cooling, 'f', 2, 64)

			if 20-len(row1) > 0 {
				row1 += strings.Repeat(" ", 20-len(row1))
			}

			if 20-len(row2) > 0 {
				row2 += strings.Repeat(" ", 20-len(row2))
			}
			msg += row1[0:20]
			msg += row2[0:20]

			_, err := serial.Port.Write([]byte(msg))
			if err != nil {
				log.Debug("Failed write to display:", err)
			}
		}
	}
} // }}}

// WORKER that monitors the current connection state
func monitorState(connection *basenode.Connection) { // {{{
	for s := range connection.State {
		switch s {
		case basenode.ConnectionStateConnected:
			connection.Send <- node.Node()
		case basenode.ConnectionStateDisconnected:
		}
	}
} // }}}

// WORKER that recives all incomming commands
func serverRecv(connection *basenode.Connection) { // {{{
	for d := range connection.Receive {
		if err := processCommand(d); err != nil {
			log.Error(err)
		}
	}
} // }}}

func processCommand(cmd protocol.Command) error { // {{{
	var target bool

	if len(cmd.Args) < 1 {
		if len(cmd.Params) < 1 {
			return fmt.Errorf("Missing arguments, ignoring command")
		} else {
			target = cmd.Params[0] != "" && cmd.Params[0] != "false"
		}
	} else {
		target = cmd.Args[0] != "" && cmd.Args[0] != "0"
	}

	switch cmd.Cmd {
	case "CirculationPumps":
		if target {
			ard.Port.Write([]byte{0x02, 0x01, 0x01, 0x03}) // Turn on
		} else {
			ard.Port.Write([]byte{0x02, 0x01, 0x00, 0x03}) // Turn off
		}
		break
	case "Skimmer":
		if target {
			ard.Port.Write([]byte{0x02, 0x02, 0x01, 0x03}) // Turn on
		} else {
			ard.Port.Write([]byte{0x02, 0x02, 0x00, 0x03}) // Turn off
		}
		break
	case "Heating":
		if target {
			ard.Port.Write([]byte{0x02, 0x03, 0x01, 0x03}) // Turn on
		} else {
			ard.Port.Write([]byte{0x02, 0x03, 0x00, 0x03}) // Turn off
		}
		break
	case "CoolingP":
		i, err := strconv.Atoi(cmd.Args[0])
		if err != nil {
			return fmt.Errorf("Failed to decode arg[0] to int %s %s", err, cmd.Args[0])
		}

		ard.Port.Write([]byte{0x02, 0x04, byte(i), 0x03}) // Turn on
		break
	case "CoolingI":
		i, err := strconv.Atoi(cmd.Args[0])
		if err != nil {
			return fmt.Errorf("Failed to decode arg[0] to int %s %s", err, cmd.Args[0])
		}

		ard.Port.Write([]byte{0x02, 0x05, byte(i), 0x03}) // Turn on
		break
	case "CoolingD":
		i, err := strconv.Atoi(cmd.Args[0])
		if err != nil {
			return fmt.Errorf("Failed to decode arg[0] to int %s %s", err, cmd.Args[0])
		}

		ard.Port.Write([]byte{0x02, 0x06, byte(i), 0x03}) // Turn on
		break
	}
	return nil
} // }}}

// Serial connection workers
func (config *SerialConnection) run(connection *basenode.Connection, callback func(data string, connection *basenode.Connection)) { // {{{
	go func() {
		for {
			config.connect(connection, callback)
			<-time.After(time.Second * 4)
		}
	}()
}                                                                                                                                       // }}}
func (config *SerialConnection) connect(connection *basenode.Connection, callback func(data string, connection *basenode.Connection)) { // {{{

	c := &serial.Config{Name: config.Name, Baud: config.Baud}
	var err error

	config.Port, err = serial.OpenPort(c)
	if err != nil {
		log.Error("Serial connect failed: ", err)
		return
	}

	var incomming string = ""

	for {
		buf := make([]byte, 128)

		n, err := config.Port.Read(buf)
		if err != nil {
			log.Error("Serial read failed: ", err)
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
			fmt.Print(msg, "\r")

			go callback(msg, connection)
		}
	}
} // }}}
