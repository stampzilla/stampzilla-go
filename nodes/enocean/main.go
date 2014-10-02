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
	send := make(chan interface{})
	recv := make(chan protocol.Command)
	connectionState := basenode.Connect(send, recv)
	go monitorState(connectionState, send)
	go serverRecv(recv)

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

			if b, ok := p.(*goenocean.TelegramRps); ok {
				eep := goenocean.NewEepF60201()
				eep.SetTelegram(b) //THIS IS COOL!

				fmt.Println("EB:", eep.EnergyBow())
				fmt.Println("R1B0:", eep.R1B0())
				fmt.Println("R2B0:", eep.R2B0())
				fmt.Println("R2B1:", eep.R2B1())
				fmt.Printf("raw data: %b\n", eep.TelegramData())
			}

		}
	}
}
