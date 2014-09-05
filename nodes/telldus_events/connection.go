package main

import (
	"encoding/json"
	"fmt"
	log "github.com/cihub/seelog"
	"github.com/stampzilla/stampzilla-go/protocol"
	"net"
	"time"
)

var Connection net.Conn

func connection(host, port string, info *protocol.Node) {
	var err error
	for {
		log.Info("Connection to ", host, ":", port)
		Connection, err = net.Dial("tcp", net.JoinHostPort(host, port))

		if err != nil {
			log.Error("Failed connection: ", err)
			<-time.After(time.Second)
			continue
		}

		log.Trace("Connected")

		connectionWorker(info)

		log.Warn("Lost connection, reconnecting")
		<-time.After(time.Second)
	}
}

func connectionWorker(info *protocol.Node) {
	// Send update
	sendUpdate(info)

	// Recive data
	for {
		buf := make([]byte, 51200)
		nr, err := Connection.Read(buf)
		if err != nil {
			return
		}

		data := buf[0:nr]

		var cmd protocol.Command
		err = json.Unmarshal(data, &cmd)
		if err != nil {
			log.Warn(err)
		} else {
			log.Info(cmd)
			processCommand(cmd)
		}

	}
}

func sendUpdate(info *protocol.Node) {
	b, err := json.Marshal(info)
	if err != nil {
		log.Error(err)
	}
	fmt.Fprintf(Connection, string(b))
}
