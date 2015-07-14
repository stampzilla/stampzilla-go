package main

import (
	"encoding/binary"
	"flag"
	"log"
	"time"

	"github.com/stampzilla/stampzilla-go/nodes/basenode"
	"github.com/stampzilla/stampzilla-go/protocol"
)

// MAIN - This is run when the init function is done
func main() {
	log.Println("Starting modbus node")

	// Parse all commandline arguments, host and port parameters are added in the basenode init function
	flag.Parse()

	//Get a config with the correct parameters
	config := basenode.NewConfig()

	//Activate the config
	basenode.SetConfig(config)

	node := protocol.NewNode("modbus")

	registers := NewRegisters()
	registers.ReadFromFile("registers.json")

	modbusConnection := &Modbus{}
	err := modbusConnection.Connect()

	if err != nil {
		log.Println(err)
		return
	}

	defer modbusConnection.Close()

	//REG_HC_TEMP_IN1 214 Reg
	//REG_HC_TEMP_IN2 215 Reg
	//REG_HC_TEMP_IN3 216 Reg
	//REG_HC_TEMP_IN4 217 Reg
	//REG_HC_TEMP_IN5 218 Reg

	//REG_DAMPER_PWM 301 Reg
	//REG_HC_WC_SIGNAL 204 Reg

	//client := modbus.NewClient(handler)
	//modbus.NewClient
	//results, _ := client.ReadHoldingRegisters(214, 1)
	//if err != nil {
	//log.Println(err)
	//}
	results, _ := modbusConnection.ReadInputRegister(214)
	log.Println("REG_HC_TEMP_IN1: ", results)
	results, _ = modbusConnection.ReadInputRegister(215)
	log.Println("REG_HC_TEMP_IN2: ", results)
	results, _ = modbusConnection.ReadInputRegister(216)
	log.Println("REG_HC_TEMP_IN3: ", results)
	results, _ = modbusConnection.ReadInputRegister(217)
	log.Println("REG_HC_TEMP_IN4: ", results)
	results, _ = modbusConnection.ReadInputRegister(218)
	log.Println("REG_HC_TEMP_IN5: ", binary.BigEndian.Uint16(results))
	results, _ = modbusConnection.ReadInputRegister(207)
	log.Println("REG_HC_TEMP_LVL: ", results)
	results, _ = modbusConnection.ReadInputRegister(301)
	log.Println("REG_DAMPER_PWM: ", results)
	results, _ = modbusConnection.ReadInputRegister(204)
	log.Println("REG_HC_WC_SIGNAL: ", results)
	results, _ = modbusConnection.ReadInputRegister(209)
	log.Println("REG_HC_TEMP_LVL1-5: ", results)
	results, _ = modbusConnection.ReadInputRegister(101)
	log.Println("100 REG_FAN_SPEED_LEVEL: ", results)

	//Start communication with the server
	connection := basenode.Connect()

	// Thit worker keeps track on our connection state, if we are connected or not

	//node.AddElement(&protocol.Element{
	//Type: protocol.ElementTypeColorPicker,
	//Name: "Example color picker",
	//Command: &protocol.Command{
	//Cmd:  "color",
	//Args: []string{"1"},
	//},
	//Feedback: "Devices[4].State",
	//})

	//state := NewState()
	node.SetState(registers)

	// This worker recives all incomming commands
	go serverRecv(registers, connection, modbusConnection)
	go monitorState(node, connection, registers, modbusConnection)
	select {}
}

// WORKER that monitors the current connection state
func monitorState(node *protocol.Node, connection *basenode.Connection, registers *Registers, modbusConnection *Modbus) {
	var stopFetching chan bool
	for s := range connection.State {
		switch s {
		case basenode.ConnectionStateConnected:
			fetchRegisters(registers, modbusConnection)
			stopFetching = periodicalFetcher(registers, modbusConnection, connection, node)
			connection.Send <- node.Node()
		case basenode.ConnectionStateDisconnected:
			close(stopFetching)
		}
	}
}

func periodicalFetcher(registers *Registers, connection *Modbus, nodeConn *basenode.Connection, node *protocol.Node) chan bool {

	ticker := time.NewTicker(30 * time.Second)
	quit := make(chan bool)
	go func() {
		for {
			select {
			case <-ticker.C:
				fetchRegisters(registers, connection)
				nodeConn.Send <- node.Node()
			case <-quit:
				ticker.Stop()
				log.Println("Stopping periodicalFetcher")
				return
			}
		}
	}()

	return quit
}

func fetchRegisters(registers *Registers, connection *Modbus) {
	for _, v := range registers.Registers {

		data, err := connection.ReadInputRegister(v.Id)
		if err != nil {
			log.Println(err)
			continue
		}
		if len(data) != 2 {
			log.Println("Wrong length, expected 2")
			continue
		}

		if v.Base != 0 {
			v.Value = float64(binary.BigEndian.Uint16(data)) / float64(v.Base)
			continue
		}
		v.Value = binary.BigEndian.Uint16(data)
	}
}

// WORKER that recives all incomming commands
func serverRecv(registers *Registers, connection *basenode.Connection, modbusConnection *Modbus) {
	for d := range connection.Receive {
		processCommand(registers, connection, d)
	}
}

// THis is called on each incomming command
func processCommand(registers *Registers, connection *basenode.Connection, cmd protocol.Command) {
	//if s, ok := node.State.(*Registers); ok {
	//log.Println("Incoming command from server:", cmd)
	//if len(cmd.Args) == 0 {
	//return
	//}
	//device := s.Device(cmd.Args[0])

	//switch cmd.Cmd {
	//case "on":
	//device.State = true
	//connection.Send <- node.Node()
	//case "off":
	//device.State = false
	//connection.Send <- node.Node()
	//case "toggle":
	//log.Println("got toggle")
	//if device.State {
	//device.State = false
	//} else {
	//device.State = true
	//}
	//connection.Send <- node.Node()
	//}
	//}
}
