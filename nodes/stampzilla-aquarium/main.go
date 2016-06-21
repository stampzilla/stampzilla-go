package main

import (
	"flag"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"time"

	"syscall"
	"unsafe"

	"math"

	"github.com/tarm/goserial"

	log "github.com/cihub/seelog"
	"github.com/stampzilla/stampzilla-go/nodes/basenode"
	"github.com/stampzilla/stampzilla-go/protocol"

	"github.com/stampzilla/stampzilla-go/pkg/notifier"
)

type SerialConnection struct {
	Name string
	Baud int
	Port io.ReadWriteCloser
}

type winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

var node *protocol.Node
var state *State = &State{}
var notify *notifier.Notify
var serverConnection basenode.Connection
var ard *SerialConnection
var disp *SerialConnection
var frameCount int
var lastTemp float64
var phFilter []float64
var waterLevelFilter []float64

var filterFilterAlarm int
var filterFillingAlarm int

var rateTimestamp int64
var rateCount int64
var rate int64

func main() { // {{{
	config := basenode.NewConfig()
	basenode.SetConfig(config)

	var arddev string
	var dispdev string
	flag.StringVar(&arddev, "arduino-dev", "/dev/arduino", "Arduino serial port")
	flag.StringVar(&dispdev, "display-dev", "/dev/poleDisplay", "Display serial port")

	flag.Parse()

	phFilter = make([]float64, 0)

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

	node.AddElement(&protocol.Element{
		Type: protocol.ElementTypeSlider,
		Name: "White",
		Command: &protocol.Command{
			Cmd:  "dim",
			Args: []string{"white"},
		},
		Feedback: "Lights.White",
	})
	node.AddElement(&protocol.Element{
		Type: protocol.ElementTypeSlider,
		Name: "Red",
		Command: &protocol.Command{
			Cmd:  "dim",
			Args: []string{"red"},
		},
		Feedback: "Lights.Red",
	})
	node.AddElement(&protocol.Element{
		Type: protocol.ElementTypeSlider,
		Name: "Green",
		Command: &protocol.Command{
			Cmd:  "dim",
			Args: []string{"green"},
		},
		Feedback: "Lights.Green",
	})
	node.AddElement(&protocol.Element{
		Type: protocol.ElementTypeSlider,
		Name: "Blue",
		Command: &protocol.Command{
			Cmd:  "dim",
			Args: []string{"blue"},
		},
		Feedback: "Lights.Blue",
	})
	node.AddElement(&protocol.Element{
		Type:     protocol.ElementTypeText,
		Name:     "Cover temperature",
		Feedback: `Lights.Temperature`,
	})

	node.AddElement(&protocol.Element{
		Type:     protocol.ElementTypeText,
		Name:     "pH probe",
		Feedback: `PH`,
	})

	node.AddElement(&protocol.Element{
		Type: protocol.ElementTypeButton,
		Name: "Dose 1 (1s)",
		Command: &protocol.Command{
			Cmd:  "dose1",
			Args: []string{"1000"},
		},
	})
	node.AddElement(&protocol.Element{
		Type: protocol.ElementTypeButton,
		Name: "Dose 2 (1s)",
		Command: &protocol.Command{
			Cmd:  "dose2",
			Args: []string{"1000"},
		},
	})
	node.AddElement(&protocol.Element{
		Type: protocol.ElementTypeButton,
		Name: "Dose 3 (1s)",
		Command: &protocol.Command{
			Cmd:  "dose3",
			Args: []string{"1000"},
		},
	})
	node.AddElement(&protocol.Element{
		Type: protocol.ElementTypeButton,
		Name: "Dose 4 (1s)",
		Command: &protocol.Command{
			Cmd:  "dose4",
			Args: []string{"1000"},
		},
	})
	node.AddElement(&protocol.Element{
		Type: protocol.ElementTypeButton,
		Name: "Dose 5 (1s)",
		Command: &protocol.Command{
			Cmd:  "dose5",
			Args: []string{"1000"},
		},
	})
	node.AddElement(&protocol.Element{
		Type: protocol.ElementTypeButton,
		Name: "Dose 6 (1s)",
		Command: &protocol.Command{
			Cmd:  "dose6",
			Args: []string{"1000"},
		},
	})

	serverConnection = basenode.Connect()
	notify = notifier.New(serverConnection)
	notify.SetSource(node)

	go monitorState(serverConnection)

	// This worker recives all incomming commands
	go serverRecv(serverConnection)

	// Setup the serial connection
	ard = &SerialConnection{Name: arddev, Baud: 115200}
	ard.run(serverConnection, func(data string, connection basenode.Connection) {
		processArduinoData(data, connection)
	})

	disp = &SerialConnection{Name: dispdev, Baud: 9600}
	disp.run(serverConnection, func(data string, connection basenode.Connection) {
		log.Debug("Incomming from display", data)
	})

	go updateDisplay(serverConnection, disp)

	/*go func() {
		for {
			fmt.Print("\r", frameCount)
			<-time.After(time.Millisecond * 10)
		}
	}()*/

	select {}
} // }}}

var updateInhibit bool = false
var changed bool = false

func processArduinoData(msg string, connection basenode.Connection) { // {{{
	//log.Debug(msg)

	var prevState State = *state

	values := strings.Split(msg, "|")
	if len(values) < 13 {
		printTerminalStatus("Invalid length")
		log.Warn("Invalid message: ", msg)
		return
	}

	// Temperature// {{{
	value, err := strconv.ParseFloat(values[0], 64)
	if err != nil {
		return
	}
	if value != state.WaterTemperature {
		changed = true
	}
	state.WaterTemperature = value // }}}
	value, err = strconv.ParseFloat(values[4], 64)
	if err != nil {
		return
	}
	lastTemp = value

	// Filling stat0e // {{{
	value, err = strconv.ParseFloat(values[1], 64)
	if err != nil {
		return
	}
	if state.FillingTime != value {
		changed = true
	}
	state.FillingTime = value // }}}

	// Cooling %// {{{
	value, err = strconv.ParseFloat(values[2], 64)
	if err != nil {
		return
	}
	if state.Cooling != value {
		changed = true
	}
	state.Cooling = value // }}}

	bits := values[3][0]
	CirculationPumps := bits&0x01 != 0
	Skimmer := bits&0x02 != 0
	Heating := bits&0x04 != 0
	Filling := bits&0x08 == 0
	WaterLevelOk := bits&0x10 == 0
	FilterOk := bits&0x20 == 0

	if state.FilterOk != FilterOk {
		if filterFilterAlarm < 500 {
			filterFilterAlarm++
		} else if filterFilterAlarm <= 500 {
			filterFilterAlarm++
			notify.Error("Filter igensatt")
		}
	} else {
		filterFilterAlarm = 0
	}

	if state.Filling != Filling && state.FillingTime > 0 && state.CirculationPumps {
		if filterFillingAlarm < 500 {
			filterFillingAlarm++
		} else if filterFillingAlarm <= 500 {
			filterFillingAlarm++
			notify.Error("Påfyllnad misslyckad")
		}
	} else {
		filterFillingAlarm = 0
	}

	state.CirculationPumps = CirculationPumps
	state.Skimmer = Skimmer
	state.Heating = Heating
	state.Filling = Filling
	state.WaterLevelOk = WaterLevelOk
	state.FilterOk = FilterOk

	light := strings.Split(values[6], "*")
	value, err = strconv.ParseFloat(light[0], 64)
	if err == nil {
		state.Lights.Red = value
	}
	value, err = strconv.ParseFloat(light[1], 64)
	if err == nil {
		state.Lights.Green = value
	}
	value, err = strconv.ParseFloat(light[2], 64)
	if err == nil {
		state.Lights.Blue = value
	}
	value, err = strconv.ParseFloat(light[3], 64)
	if err == nil {
		state.Lights.White = value
	}
	value, err = strconv.ParseFloat(values[7], 64)
	if err == nil {
		if value > 0 {
			state.Lights.Temperature = value
		}
	} else {
		state.Lights.Temperature = -1
	}

	pH := strings.Split(values[8], ":")
	value, err = strconv.ParseFloat(pH[0], 64)
	if err == nil {
		ph := ((673-value)/828*14+7)*1.5650273224 - 3.84
		if len(phFilter) < 200 {
			phFilter = append(phFilter, ph)
		} else {
			phFilter = append(phFilter[1:], ph)
			state.PH = toFixed(Average(phFilter), 2)
		}
	}

	air := strings.Split(values[9], ",")
	if len(air) == 2 {
		value, err = strconv.ParseFloat(air[0], 64)
		if err == nil && value <= 100 {
			state.Humidity = value
		}
		value, err = strconv.ParseFloat(air[1], 64)
		if err == nil && value > 0 && value < 100 {
			state.AirTemperature = value
		}
	}

	value, err = strconv.ParseFloat(values[10], 64)
	if err == nil {
		if value > 0 {
			state.Lights.Cooling = value
		}
	} else {
		state.Lights.Cooling = -1
	}

	value, err = strconv.ParseFloat(values[11], 64)
	if err == nil {
		waterLevel := toFixed((value-497)/313*213, 1)
		if len(waterLevelFilter) < 200 {
			waterLevelFilter = append(waterLevelFilter, waterLevel)
		} else {
			waterLevelFilter = append(waterLevelFilter[1:], waterLevel)
			state.WaterLevel = toFixed(Average(waterLevelFilter), 0)
		}
	} else {
		state.WaterLevel = -1
	}

	state.Error, err = strconv.Atoi(values[12])
	if prevState.Error != state.Error {
		log.Critical("Error detected: !", state.Error)
		switch state.Error {
		case 0:
			notify.Info("Errors was restored..")
		case 1:
			notify.Error("Low water temperature")
		case 2:
			notify.Error("High water temperature")
		case 3:
			notify.Error("Topup pump exeeded maximum run time")
		default:
			notify.Error("SYSTEM ERROR?! Super suspicious unknown error!?")
		}
	}

	// Check if something have changed
	if !reflect.DeepEqual(prevState, *state) {
		changed = true
	}

	frameCount++
	rateCount++

	if time.Now().Unix() != rateTimestamp {
		rateTimestamp = time.Now().Unix()
		rate = rateCount
		rateCount = 0
	}

	//fmt.Print(frameCount, " - ", msg, " - ", bits, "\r")
	printTerminalStatus(strconv.Itoa(int(rate)) + " msg/s " + msg + "        " + strconv.FormatFloat(state.WaterLevel, 'g', -1, 64))

	if !updateInhibit && changed {
		changed = false
		//fmt.Print("\n")

		connection.Send(node.Node())

		updateInhibit = true
		//log.Warn("Invalid pacakge: ", msg)
		go func() {
			<-time.After(time.Millisecond * 200)
			updateInhibit = false
		}()
	}
} // }}}

func Average(xs []float64) float64 {
	total := float64(0)
	for _, x := range xs {
		total += x
	}
	return total / float64(len(xs))
}
func updateDisplay(connection basenode.Connection, serial *SerialConnection) { // {{{

	serial.Port.Write([]byte("\x1B")) // reset

	for {
		select {
		case <-time.After(time.Second / 15):
			if serial.Port == nil {
				continue
			}

			var msg, row1, row2 string
			var wLevel string = "OK"

			if lastTemp == 0 {
				wLevel = "NO TEMP"
			}

			//if !state.WaterLevelOk {
			//wLevel = "WLVL"
			//}
			if state.FillingTime > 0 {
				wLevel = "FILL ERR"
				if state.FillingTime < 20000 {
					wLevel = "FILLING"
				}
			}

			if !state.FilterOk {
				wLevel = "FILTER!"
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
func monitorState(connection basenode.Connection) { // {{{
	for s := range connection.State() {
		switch s {
		case basenode.ConnectionStateConnected:
			connection.Send(node.Node())
		case basenode.ConnectionStateDisconnected:
		}
	}
} // }}}

// WORKER that recives all incomming commands
func serverRecv(connection basenode.Connection) { // {{{
	for d := range connection.Receive() {
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
			ard.Port.Write([]byte{0x02, 0x01, 0x00, 0x01, 0x03}) // Turn on
		} else {
			ard.Port.Write([]byte{0x02, 0x01, 0x00, 0x00, 0x03}) // Turn off
		}
		break
	case "Skimmer":
		if target {
			ard.Port.Write([]byte{0x02, 0x02, 0x00, 0x01, 0x03}) // Turn on
		} else {
			ard.Port.Write([]byte{0x02, 0x02, 0x00, 0x00, 0x03}) // Turn off
		}
		break
	case "Heating":
		if target {
			ard.Port.Write([]byte{0x02, 0x03, 0x00, 0x01, 0x03}) // Turn on
		} else {
			ard.Port.Write([]byte{0x02, 0x03, 0x00, 0x00, 0x03}) // Turn off
		}
		break
	case "CoolingP":
		i, err := strconv.Atoi(cmd.Args[0])
		if err != nil {
			return fmt.Errorf("Failed to decode arg[0] to int %s %s", err, cmd.Args[0])
		}

		ard.Port.Write([]byte{0x02, 0x04, 0x00, byte(i), 0x03}) // Turn on
		break
	case "CoolingI":
		i, err := strconv.Atoi(cmd.Args[0])
		if err != nil {
			return fmt.Errorf("Failed to decode arg[0] to int %s %s", err, cmd.Args[0])
		}

		ard.Port.Write([]byte{0x02, 0x05, 0x00, byte(i), 0x03}) // Turn on
		break
	case "CoolingD":
		i, err := strconv.Atoi(cmd.Args[0])
		if err != nil {
			return fmt.Errorf("Failed to decode arg[0] to int %s %s", err, cmd.Args[0])
		}

		ard.Port.Write([]byte{0x02, 0x06, 0x00, byte(i), 0x03}) // Turn on
		break
	case "dim":
		var i int
		var err error

		switch {
		case len(cmd.Args) == 1:
			i, err = strconv.Atoi(cmd.Params[0])
			if err != nil {
				return fmt.Errorf("Failed to decode param[0] to int %s %s", err, cmd.Args[0])
			}
		case len(cmd.Args) == 2:
			i, err = strconv.Atoi(cmd.Args[1])
			if err != nil {
				return fmt.Errorf("Failed to decode arg[1] to int %s %s", err, cmd.Args[0])
			}

		}

		switch cmd.Args[0] {
		case "red":
			ard.Port.Write([]byte{0x02, 0x07, 0x00, byte(i), 0x03}) // Turn on
		case "green":
			ard.Port.Write([]byte{0x02, 0x08, 0x00, byte(i), 0x03}) // Turn on
		case "blue":
			ard.Port.Write([]byte{0x02, 0x09, 0x00, byte(i), 0x03}) // Turn on
		case "white":
			ard.Port.Write([]byte{0x02, 0x0A, 0x00, byte(i), 0x03}) // Turn on
		}
	case "cooling":
		i, err := strconv.Atoi(cmd.Args[0])
		if err != nil {
			return fmt.Errorf("Failed to decode arg[0] to int %s %s", err, cmd.Args[0])
		}

		ard.Port.Write([]byte{0x02, 0x0B, 0x00, byte(i), 0x03}) // Turn on
		break

	case "dose1":
		i, err := strconv.Atoi(cmd.Args[0])
		if err != nil {
			return fmt.Errorf("Failed to decode arg[0] to int %s %s", err, cmd.Args[0])
		}

		ard.Port.Write([]byte{0x02, 12, byte(i >> 8), byte(i), 0x03}) // command dose
		break
	case "dose2":
		i, err := strconv.Atoi(cmd.Args[0])
		if err != nil {
			return fmt.Errorf("Failed to decode arg[0] to int %s %s", err, cmd.Args[0])
		}

		ard.Port.Write([]byte{0x02, 13, byte(i >> 8), byte(i), 0x03}) // command dose
		break
	case "dose3":
		i, err := strconv.Atoi(cmd.Args[0])
		if err != nil {
			return fmt.Errorf("Failed to decode arg[0] to int %s %s", err, cmd.Args[0])
		}

		ard.Port.Write([]byte{0x02, 14, byte(i >> 8), byte(i), 0x03}) // command dose
		break
	case "dose4":
		i, err := strconv.Atoi(cmd.Args[0])
		if err != nil {
			return fmt.Errorf("Failed to decode arg[0] to int %s %s", err, cmd.Args[0])
		}

		ard.Port.Write([]byte{0x02, 15, byte(i >> 8), byte(i), 0x03}) // command dose
		break
	case "dose5":
		i, err := strconv.Atoi(cmd.Args[0])
		if err != nil {
			return fmt.Errorf("Failed to decode arg[0] to int %s %s", err, cmd.Args[0])
		}

		ard.Port.Write([]byte{0x02, 16, byte(i >> 8), byte(i), 0x03}) // command dose
		break
	case "dose6":
		i, err := strconv.Atoi(cmd.Args[0])
		if err != nil {
			return fmt.Errorf("Failed to decode arg[0] to int %s %s", err, cmd.Args[0])
		}

		ard.Port.Write([]byte{0x02, 17, byte(i >> 8), byte(i), 0x03}) // command dose
		break
	}
	return nil
} // }}}

func printTerminalStatus(msg string) { // {{{
	size := getWindowSize()

	if size.Col < 20 {
		return
	}

	if len(msg) > int(size.Col)-1 {
		msg = msg[:size.Col]
	}

	log.Flush()

	// Header
	fmt.Println(strings.Repeat(" ", int(size.Col)-1))
	fmt.Println(strings.Repeat("-", int(size.Col)-1))

	// Raw message
	printWithLimit(msg, int(size.Col))

	// Framecount
	msg = "Framecount: " + strconv.Itoa(frameCount)
	printWithLimit(msg, int(size.Col))

	fmt.Print("\033[4A\r") // Move cursor up 2 lines
}                                             // }}}
func printWithLimit(msg string, length int) { // {{{
	pad := length - len(msg) - 1

	if pad < 1 {
		fmt.Print(msg[:len(msg)+pad])
	} else {
		fmt.Print(msg + strings.Repeat(" ", pad) + "\n")
	}
}                               // }}}
func getWindowSize() *winsize { // {{{
	ws := &winsize{}
	retCode, _, _ := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(ws)))

	if int(retCode) == -1 {
		ws.Col = 0
	}
	return ws
} // }}}

// Serial connection workers
func (config *SerialConnection) run(connection basenode.Connection, callback func(data string, connection basenode.Connection)) { // {{{
	connected := make(chan struct{})

	go func() {
		for {
			config.connect(connection, callback, connected)
			<-time.After(time.Second * 4)
			connected = make(chan struct{})
		}
	}()

	<-connected
} // }}}

func (config *SerialConnection) connect(connection basenode.Connection, callback func(data string, connection basenode.Connection), connected chan struct{}) { // {{{
	red := state.Lights.Red
	green := state.Lights.Green
	blue := state.Lights.Blue
	white := state.Lights.White
	cooling := state.Cooling

	c := &serial.Config{Name: config.Name, Baud: config.Baud}
	var err error

	config.Port, err = serial.OpenPort(c)
	if err != nil {
		//log.Error("Serial connect failed: ", err)
		return
	}

	close(connected)
	<-time.After(time.Second)

	config.Port.Write([]byte{0x02, 0x07, 0x00, byte(red), 0x03})     // red
	config.Port.Write([]byte{0x02, 0x08, 0x00, byte(green), 0x03})   // green
	config.Port.Write([]byte{0x02, 0x09, 0x00, byte(blue), 0x03})    // blue
	config.Port.Write([]byte{0x02, 0x0A, 0x00, byte(white), 0x03})   // white
	config.Port.Write([]byte{0x02, 0x0B, 0x00, byte(cooling), 0x03}) // cooling

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

			go callback(msg, connection)
		}
	}
} // }}}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func toFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}
