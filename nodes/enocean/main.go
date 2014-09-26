// Package main provides ...
package main

import (
	"fmt"

	"github.com/jonaz/goenocean"
)

func main() {

	send := make(chan goenocean.PacketInterface)
	recv := make(chan goenocean.PacketInterface)
	goenocean.Serial(send, recv)
	reciever(recv)
}

func reciever(recv chan goenocean.PacketInterface) {
	for {
		select {
		case p := <-recv:
			fmt.Printf("% x\n", p)
			fmt.Printf("Packet\t %+v\n", p)
			fmt.Printf("Header\t %+v\n", p.Header())
			fmt.Printf("senderID: % x\n", p.SenderId())

			if b, ok := p.(*goenocean.PacketEepRps); ok {
				fmt.Printf("Action: %d\n", b.Action())
				fmt.Printf("Action: %b\n", b.Action())
				fmt.Printf("Action: %s\n", b.Action())
			}
		}
	}
}
