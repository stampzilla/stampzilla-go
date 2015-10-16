package main

import (
	"bufio"
	"fmt"
	"net"
	"time"

	log "github.com/cihub/seelog"
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
		response: make(chan struct{}),
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
	log.Info("Sending: ", cmd)
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
		log.Error("We got TIMOUT after send")
		c <- struct{}{}
		return
	}

}

func (s *squeezebox) writer() {
	for v := range s.write {
		_, err := s.conn.Write([]byte(v))
		if err != nil {
			log.Error(err)
		}
	}
}

func (s *squeezebox) reader() {
	connbuf := bufio.NewReader(s.conn)
	for {
		str, err := connbuf.ReadString('\n')
		if err != nil {
			log.Error(err)
			break
		}
		if len(str) > 0 {
			select {
			case s.response <- struct{}{}:
			default:
			}
			s.read <- str
		}
	}
}
