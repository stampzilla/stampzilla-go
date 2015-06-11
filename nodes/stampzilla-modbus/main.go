package main

import (
	"flag"
	"time"

	log "github.com/cihub/seelog"
	"github.com/goburrow/modbus"
	"github.com/stampzilla/stampzilla-go/nodes/basenode"
	"github.com/stampzilla/stampzilla-go/protocol"
)

// MAIN - This is run when the init function is done
func main() { /*{{{*/
	log.Info("Starting SIMPLE node")

	// Parse all commandline arguments, host and port parameters are added in the basenode init function
	flag.Parse()

	//Get a config with the correct parameters
	config := basenode.NewConfig()

	//Activate the config
	basenode.SetConfig(config)

	node := protocol.NewNode("simple")

	// Modbus RTU/ASCII
	handler := modbus.NewRTUClientHandler("/dev/ttyUSB0")
	handler.BaudRate = 9600
	handler.DataBits = 8
	handler.Parity = "E"
	handler.StopBits = 1
	handler.SlaveId = 1
	handler.Timeout = 5 * time.Second

	err := handler.Connect()
	if err != nil {
		log.Error(err)
		return
	}

	defer handler.Close()

	//REG_HC_TEMP_IN1 214 Reg
	//REG_HC_TEMP_IN2 215 Reg
	//REG_HC_TEMP_IN3 216 Reg
	//REG_HC_TEMP_IN4 217 Reg
	//REG_HC_TEMP_IN5 218 Reg

	//REG_DAMPER_PWM 301 Reg
	//REG_HC_WC_SIGNAL 204 Reg

	client := modbus.NewClient(handler)
	results, err := client.ReadInputRegisters(214, 1)
	if err != nil {
		log.Error(err)
		return
	}

	log.Info("REG_HC_TEMP_IN1: ", results)

	return
	//Start communication with the server
	connection := basenode.Connect()

	// Thit worker keeps track on our connection state, if we are connected or not
	go monitorState(node, connection)

	//node.AddElement(&protocol.Element{
	//Type: protocol.ElementTypeColorPicker,
	//Name: "Example color picker",
	//Command: &protocol.Command{
	//Cmd:  "color",
	//Args: []string{"1"},
	//},
	//Feedback: "Devices[4].State",
	//})

	state := NewState()
	node.SetState(state)

	state.AddDevice("1", "Dev1", false)
	state.AddDevice("2", "Dev2", true)

	// This worker recives all incomming commands
	go serverRecv(node, connection)
	select {}
} /*}}}*/

// WORKER that monitors the current connection state
func monitorState(node *protocol.Node, connection *basenode.Connection) {
	for s := range connection.State {
		switch s {
		case basenode.ConnectionStateConnected:
			connection.Send <- node.Node()
		case basenode.ConnectionStateDisconnected:
		}
	}
}

// WORKER that recives all incomming commands
func serverRecv(node *protocol.Node, connection *basenode.Connection) {
	for d := range connection.Receive {
		processCommand(node, connection, d)
	}
}

// THis is called on each incomming command
func processCommand(node *protocol.Node, connection *basenode.Connection, cmd protocol.Command) {
	if s, ok := node.State.(*State); ok {
		log.Info("Incoming command from server:", cmd)
		if len(cmd.Args) == 0 {
			return
		}
		device := s.Device(cmd.Args[0])

		switch cmd.Cmd {
		case "on":
			device.State = true
			connection.Send <- node.Node()
		case "off":
			device.State = false
			connection.Send <- node.Node()
		case "toggle":
			log.Info("got toggle")
			if device.State {
				device.State = false
			} else {
				device.State = true
			}
			connection.Send <- node.Node()
		}
	}
}
