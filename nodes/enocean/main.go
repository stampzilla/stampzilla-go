// Package main provides ...
package main

import (
	"fmt"

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
	p.SetSenderId([4]byte{0xfe, 0xfe, 0x74, 0x9b}) //the hardcoded senderid of my PTM215 button
	//p.SetTelegramData(0x70) //off
	p.SetTelegramData(0x50) //on
	//p.SetStatus(0x30)

	fmt.Printf("sending: % x\n", p.Encode())
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
