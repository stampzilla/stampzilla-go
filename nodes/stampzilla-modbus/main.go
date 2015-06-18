package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/goburrow/modbus"
	"github.com/stampzilla/stampzilla-go/nodes/basenode"
	"github.com/stampzilla/stampzilla-go/protocol"
)

// MAIN - This is run when the init function is done
func main() {
	log.Println("Starting SIMPLE node")

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
	handler.Parity = "N"
	handler.StopBits = 2
	handler.SlaveId = 1
	handler.Timeout = 10 * time.Second
	handler.Logger = log.New(os.Stdout, "test: ", log.LstdFlags)

	err := handler.Connect()
	if err != nil {
		log.Println(err)
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
	//results, _ := client.ReadHoldingRegisters(214, 1)
	//if err != nil {
	//log.Println(err)
	//}
	results, _ := client.ReadInputRegisters(214, 1)
	log.Println("REG_HC_TEMP_IN1: ", results)
	results, _ = client.ReadInputRegisters(215, 1)
	log.Println("REG_HC_TEMP_IN2: ", results)
	results, _ = client.ReadInputRegisters(216, 1)
	log.Println("REG_HC_TEMP_IN3: ", results)
	results, _ = client.ReadInputRegisters(217, 1)
	log.Println("REG_HC_TEMP_IN4: ", results)
	results, _ = client.ReadInputRegisters(218, 1)
	log.Println("REG_HC_TEMP_IN5: ", results)
	results, _ = client.ReadInputRegisters(207, 1)
	log.Println("REG_HC_TEMP_LVL: ", results)
	results, _ = client.ReadInputRegisters(301, 1)
	log.Println("REG_DAMPER_PWM: ", results)
	results, _ = client.ReadInputRegisters(204, 1)
	log.Println("REG_HC_WC_SIGNAL: ", results)
	results, _ = client.ReadInputRegisters(209, 5)
	log.Println("REG_HC_TEMP_LVL1-5: ", results)

	time.Sleep(time.Second * 1)
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
}

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
		log.Println("Incoming command from server:", cmd)
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
			log.Println("got toggle")
			if device.State {
				device.State = false
			} else {
				device.State = true
			}
			connection.Send <- node.Node()
		}
	}
}
