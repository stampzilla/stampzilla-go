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

	//goenocean.Read(s, processPacket)
	reciever(recv)

}

func reciever(recv chan *goenocean.Packet) {
	for {
		select {
		case p := <-recv:
			fmt.Printf("Packet\t %+v\n", p)
			fmt.Printf("Header\t %+v\n", p.Header())
		}
	}
}

func processPacket(data []byte) {
	fmt.Println("raw data")
	fmt.Println(data)
	p, err := goenocean.Decode(data)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(p)
	fmt.Printf("% x\n", p)
	fmt.Printf("Packet\t %+v\n", p)
	fmt.Printf("Header\t %+v\n", p.Header())
}
