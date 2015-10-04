package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"time"
)

type squeezebox struct {
	conn     net.Conn
	read     chan string
	write    chan string
	response chan struct{}
}

func NewSqueezebox() *squeezebox {

	return &squeezebox{
		read:     make(chan string, 100),
		write:    make(chan string, 100),
		response: make(chan struct{}, 100),
	}

}

func (s *squeezebox) Read() chan string {
	return s.read
}

func (s *squeezebox) Connect(host, username, password string) (err error) {
	s.conn, err = net.Dial("tcp", host)
	if err != nil {
		return err
	}
	go s.writer()
	go s.reader()

	s.Sendf("login %s %s", username, password)
	return
}
func (s *squeezebox) SendTo(d *Device, cmd string) {
	s.Send(d.IdUrlEncoded() + " " + cmd)
}

func (s *squeezebox) Send(cmd string) {
	log.Println("Sending: ", cmd)
	wait := make(chan struct{})
	go s.waitForResponse(wait)
	s.write <- (cmd) + "\n"
	//s.write <- url.QueryEscape(cmd) + "\n"
	<-wait
}
func (s *squeezebox) Sendf(format string, a ...interface{}) {
	s.Send(fmt.Sprintf(format, a...))
}

func (s *squeezebox) waitForResponse(c chan struct{}) {
	select {
	case <-s.response:
		c <- struct{}{}
		return
	case <-time.After(time.Second * 1):
		log.Println("We got TIMOUT after send")
		c <- struct{}{}
		return
	}

}

func (s *squeezebox) writer() {
	for v := range s.write {
		s.conn.Write([]byte(v))
	}
}

func (s *squeezebox) reader() {
	connbuf := bufio.NewReader(s.conn)
	for {
		str, err := connbuf.ReadString('\n')
		if len(str) > 0 {
			s.response <- struct{}{}
			s.read <- str
		}
		if err != nil {
			log.Println(err)
			break
		}
	}
}
