// Package main provides ...
package main

import (
	"fmt"

	"github.com/jonaz/goenocean"
)

func main() {

	send := make(chan *goenocean.Packet)
	recv := make(chan *goenocean.Packet)
	goenocean.Serial(send, recv)
	reciever(recv)
}

func reciever(recv chan *goenocean.Packet) {
	for {
		select {
		case p := <-recv:
			//fmt.Printf("% x\n", p)
			fmt.Printf("Packet\t %+v\n", p)
			fmt.Printf("Header\t %+v\n", p.Header())
		}
	}
}
