// Package main provides ...
package main

import (
	"flag"
	"fmt"
	"time"

	log "github.com/cihub/seelog"
	"github.com/jonaz/goenocean"
	"github.com/stampzilla/stampzilla-go/nodes/basenode"
	"github.com/stampzilla/stampzilla-go/protocol"
)

var node *protocol.Node
var state *State
var serverSendChannel chan interface{}
var serverRecvChannel chan protocol.Command

func init() {
	var host string
	var port string
	flag.StringVar(&host, "host", "192.168.13.2", "Server host/ip")
	flag.StringVar(&port, "port", "8282", "Server port")

	flag.Parse()

	//Setup Config
	config := &basenode.Config{
		Host: host,
		Port: port}

	basenode.SetConfig(config)

	//Start communication with the server
	serverSendChannel = make(chan interface{})
	serverRecvChannel = make(chan protocol.Command)
	connectionState := basenode.Connect(serverSendChannel, serverRecvChannel)
	go monitorState(connectionState, serverSendChannel)
	go serverRecv(serverRecvChannel)

}

func main() {

	node = protocol.NewNode("enocean")
	// Describe available actions
	node.AddAction("set", "Set", []string{"Devices.Id"})
	node.AddAction("toggle", "Toggle", []string{"Devices.Id"})
	node.AddAction("dim", "Dim", []string{"Devices.Id", "value"})

	// Describe available layouts
	node.AddLayout("1", "switch", "toggle", "Devices", []string{"on"}, "Switches")
	node.AddLayout("2", "slider", "dim", "Devices", []string{"dim"}, "Dimmers")
	node.AddLayout("3", "slider", "dim", "Devices", []string{"dim"}, "Specials")

	state = NewState()
	//state.AddDevice([4]byte{1, 2, 3, 4}, "Testdevice", []string{"asdf"}, "off")
	node.SetState(state)

	setupEnoceanCommunication()
}

func monitorState(connectionState chan int, send chan interface{}) {
	for s := range connectionState {
		switch s {
		case basenode.ConnectionStateConnected:
			send <- node
		case basenode.ConnectionStateDisconnected:
		}
	}
}

func serverRecv(recv chan protocol.Command) {

	for d := range recv {
		processCommand(d)
	}

}

func processCommand(cmd protocol.Command) {
	log.Error("INCOMING COMMAND", cmd)
}

func setupEnoceanCommunication() {
	send := make(chan goenocean.Packet)
	recv := make(chan goenocean.Packet)
	goenocean.Serial(send, recv)

	go testSend(send)
	reciever(recv)
}

func testSend(send chan goenocean.Packet) {
	p := goenocean.NewTelegramRps()
	p.SetTelegramData(0x50) //on
	//p.SetStatus(0x30) //testing shows this does not need to be set! Status defaults to 0

	fmt.Println("Sending:", p.Encode())
	send <- p

	time.Sleep(time.Second * 3)
	p.SetTelegramData(0x70) //off
	send <- p
}

func reciever(recv chan goenocean.Packet) {
	for {
		select {
		case p := <-recv:
			fmt.Printf("% x\n", p)
			fmt.Printf("Packet\t %+v\n", p)
			fmt.Printf("Header\t %+v\n", p.Header())
			fmt.Printf("senderID: % x\n", p.SenderId())

			incomingPacket(p)

			if b, ok := p.(*goenocean.TelegramRps); ok {
				eep := goenocean.NewEepF60201()
				eep.SetTelegram(b) //THIS IS COOL!

				fmt.Println("EB:", eep.EnergyBow())
				fmt.Println("R1B0:", eep.R1B0())
				fmt.Println("R2B0:", eep.R2B0())
				fmt.Println("R2B1:", eep.R2B1())
				fmt.Printf("raw data: %b\n", eep.TelegramData())
			}
			//if b, ok := p.(*goenocean.TelegramVld); ok {
			//eep := goenocean.NewEepF60201()
			//eep.SetTelegram(b) //THIS IS COOL!
			//}

		}
	}
}

func incomingPacket(p goenocean.Packet) {
	d := NewDevice(p.SenderId(), "Unknown", "", "", nil)

	if val, ok := state.Devices[d.Id()]; ok {
		fmt.Println("deice already exists", val)

		if b, ok := p.(*goenocean.TelegramVld); ok {
			fmt.Println("VLD TELEGRAM DETECTED")
			eep := goenocean.NewEepD20109()
			eep.SetTelegram(b) //THIS IS COOL!
			fmt.Println(eep.CommandId())

			if eep.CommandId() == 4 {
				fmt.Println("OUTPUTVALUE", eep.OutputValue())

				if eep.OutputValue() > 0 {
					state.Devices[d.Id()].State = "ON"
					//d.State = "ON"
				} else {
					//d.State = "OFF"
					state.Devices[d.Id()].State = "OFF"
				}

				serverSendChannel <- node
			}
		}
		if b, ok := p.(*goenocean.Telegram4bs); ok {
			fmt.Println("4BS TELEGRAM DETECTED")
			eep := goenocean.NewEepA51201()
			eep.SetTelegram(b) //THIS IS COOL!
			fmt.Println("READING", eep.MeterReading())
			fmt.Println("TARIFF", eep.TariffInfo())
			fmt.Println("Datatype", eep.DataType())

			state.Devices[d.Id()].Power = eep.MeterReading()
			state.Devices[d.Id()].PowerUnit = eep.DataType()
			serverSendChannel <- node
		}
		return
	}

	//Add unknown device
	state.Devices[d.Id()] = d
	serverSendChannel <- node
}
