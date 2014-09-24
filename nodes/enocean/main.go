// Package main provides ...
package main

import (
	"fmt"
	"log"

	"github.com/jonaz/goenocean"
	"github.com/tarm/goserial"
)

func main() {

	c := &serial.Config{Name: "/dev/ttyUSB0", Baud: 57600}
	s, err := serial.OpenPort(c)
	if err != nil {
		log.Fatal(err)
	}

	//buf := make([]byte, 5000)
	//s.Read(buf)
	//fmt.Println(buf)

	//return

	goenocean.Read(s, processPacket)

}

func processPacket(data []byte) {
	fmt.Println("raw data")
	fmt.Println(data)
	p, err := goenocean.Decode(data)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(p)
	fmt.Printf("% x\n", p)
	fmt.Printf("Packet\t %+v\n", p)
	fmt.Printf("Header\t %+v\n", p.Header())
}
