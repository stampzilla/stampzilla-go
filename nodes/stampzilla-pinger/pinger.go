package main

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/stampzilla/stampzilla-go/nodes/basenode"
	"github.com/tatsushid/go-fastping"
)

type Target struct {
	Name   string
	Ip     string
	Online bool
	Ping   float64

	shutdown chan bool
	waiting  bool
	sync.Mutex
}

func (t *Target) start(connection basenode.Connection) {
	go t.worker(connection)
}

func (t *Target) stop() {
	select {
	case <-t.shutdown:
	default:
		if t.shutdown != nil {
			close(t.shutdown)
		}
	}
}

func (t *Target) worker(connection basenode.Connection) error {
	t.shutdown = make(chan bool)

	ra, err := net.ResolveIPAddr("ip4:icmp", t.Ip)
	if err != nil {
		return err
	}

	p := fastping.NewPinger()
	p.MaxRTT = time.Second - (time.Millisecond * 100)
	p.AddIPAddr(ra)

	p.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		t.Lock()
		t.waiting = false
		if !t.Online {
			fmt.Printf("Online: %s receive, RTT: %v\n", addr.String(), rtt)
			t.Online = true
		}

		t.Ping = float64(rtt) / float64(time.Millisecond)
		connection.Send(node.Node())

		t.Unlock()
	}

	p.OnIdle = func() {
		if t.waiting && t.Online {
			t.Lock()
			t.Online = false
			t.Ping = 0
			t.Unlock()

			fmt.Printf("Offline: %s\n", ra.String())
			connection.Send(node.Node())
		}
	}

	for {
		t.Lock()
		t.waiting = true
		t.Unlock()

		err = p.Run()
		if err != nil {
			return err
		}
		<-time.After(time.Second)
	}

	return nil
}
