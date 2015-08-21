package basenode

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	log "github.com/cihub/seelog"
	"github.com/stampzilla/stampzilla-go/protocol"
)

const (
	ConnectionStateConnected    = 1
	ConnectionStateDisconnected = 0
)

type Connection struct {
	Send    chan interface{}
	Receive chan protocol.Command
	State   chan int
}

func Connect() *Connection {

	connection := &Connection{
		Send:    make(chan interface{}, 100),
		Receive: make(chan protocol.Command, 100),
		State:   make(chan int),
	}

	go func() {
		for {
			quit := make(chan bool)
			log.Info("Connection to ", config.Host, ":", config.Port)
			tcpConnection, err := net.Dial("tcp", net.JoinHostPort(config.Host, config.Port))
			if err != nil {
				log.Error("Failed connection: ", err)
				<-time.After(time.Second)
				continue
			}

			connection.State <- ConnectionStateConnected
			log.Trace("Connected")
			serverIsAlive := make(chan bool)
			go timeoutMonitor(tcpConnection, serverIsAlive)
			go sendWorker(tcpConnection, connection.Send, quit)

			connectionWorker(tcpConnection, connection.Receive, serverIsAlive)
			close(quit)
			connection.State <- ConnectionStateDisconnected

			log.Warn("Lost connection, reconnecting")
			<-time.After(time.Second)
		}
	}()
	return connection
}

func sendWorker(connection net.Conn, send chan interface{}, quit chan bool) {
	var err error
	encoder := json.NewEncoder(connection)
	for {
		select {
		case d := <-send:

			if a, ok := d.(*protocol.Node); ok {
				a.SetUuid(config.Uuid)
				err = encoder.Encode(a.Node())
			} else {
				err = encoder.Encode(d)
			}
			if err != nil {
				fmt.Println("Error encoder.Encode: ", err)
			}
		case <-quit:
			log.Trace("sendWorker disconnected")
			return

		}
	}
}

func connectionWorker(connection net.Conn, recv chan protocol.Command, serverIsAlive chan bool) {
	// Recive data
	decoder := json.NewDecoder(connection)
	for {
		var cmd protocol.Command
		err := decoder.Decode(&cmd)

		if err != nil {
			if err.Error() == "EOF" {
				log.Error("EOF:", err)
				return
			}
			log.Warn(err)
			return
		} else {
			serverIsAlive <- true

			if cmd.Ping {
				connection.Write([]byte("{\"Pong\":true}"))
				continue
			}

			log.Debug("Command from server", cmd)
			recv <- cmd
		}

	}
}

func timeoutMonitor(connection net.Conn, serverIsAlive chan bool) {
	for {
		select {
		case <-serverIsAlive:
			// Everything is great, just continue
			continue
		case <-time.After(time.Second * 15):
			log.Warn("Server connection timeout, closing connection")
			connection.Close()
			return
		}
	}
}
