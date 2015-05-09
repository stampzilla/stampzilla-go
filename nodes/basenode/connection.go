package basenode

import (
	"bufio"
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
			go sendWorker(tcpConnection, connection.Send, quit)

			connectionWorker(tcpConnection, connection.Receive)
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
	var ret []byte
	for {
		select {
		case d := <-send:

			if a, ok := d.(*protocol.Node); ok {
				a.Uuid = config.Uuid
				ret, err = json.Marshal(a.Node())
			} else {
				ret, err = json.Marshal(d)
			}
			if err != nil {
				fmt.Println("Error marshal json", err)
			}
			fmt.Fprintf(connection, string(ret))
		case <-quit:
			return

		}
	}
}

func connectionWorker(connection net.Conn, recv chan protocol.Command) {
	// Recive data
	for {
		reader := bufio.NewReader(connection)
		decoder := json.NewDecoder(reader)
		var cmd protocol.Command
		err := decoder.Decode(&cmd)

		//err = json.Unmarshal(data, &cmd)
		if err != nil {
			if err.Error() == "EOF" {
				return
			}
			log.Warn(err)
			//return here?
		} else {
			log.Debug("Command from server", cmd)
			recv <- cmd
		}

	}
}
