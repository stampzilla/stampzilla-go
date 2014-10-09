package main

import (
	"encoding/json"
	"fmt"
	"net"

	log "github.com/cihub/seelog"
	"github.com/stampzilla/stampzilla-go/protocol"
)

var NodesConnection map[string]net.Conn
var NodesWait map[string]chan bool

func netStart(port string) {
	l, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Println("listen error", err)
		return
	}

	NodesConnection = make(map[string]net.Conn)
	NodesWait = make(map[string]chan bool)

	go func() {
		for {
			fd, err := l.Accept()
			if err != nil {
				fmt.Println("accept error", err)
				return
			}

			go newClient(fd)
		}
	}()
}

// Handle a client
func newClient(c net.Conn) {
	log.Info("New client connected")
	id := ""
	for {
		buf := make([]byte, 51200)
		nr, err := c.Read(buf)
		if err != nil {
			log.Info(id, " - Client disconnected")
			if id != "" {
				delete(Nodes, id)
			}
			//TODO be able to not send everything always.
			clients.messageOtherClients(&Message{"all", Nodes})
			return
		}

		//TODO: Handle when multiple messages gets concated ex: msg}{msg2
		data := buf[0:nr]

		var info protocol.Node
		err = json.Unmarshal(data, &info)
		if err != nil {
			log.Warn(err, " -->", string(data), "<--")
		} else {
			id = info.Id
			Nodes[info.Id] = info
			NodesConnection[info.Id] = c

			log.Info(info.Id, " - Got update on state")

			if NodesWait[info.Id] != nil {
				select {
				case NodesWait[info.Id] <- false:
					close(NodesWait[info.Id])
					NodesWait[info.Id] = nil
				default:
				}
			}

			clients.messageOtherClients(&Message{"singlenode", Nodes[info.Id]})
		}
	}
}
