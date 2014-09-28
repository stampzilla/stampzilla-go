// Package main provides ...
package main

import (
	"fmt"
	"time"

	"github.com/jonaz/goenocean"
)

func main() {

	send := make(chan goenocean.Packet)
	recv := make(chan goenocean.Packet)
	goenocean.Serial(send, recv)

	testSend(send)
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
				fmt.Printf("Action: %d\n", b.Action())
				fmt.Printf("Action: %b\n", b.Action())
				fmt.Printf("Action: %s\n", b.Action())
			}

		}
	}
}
