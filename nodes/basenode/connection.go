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

func Connect(send chan interface{}, recv chan protocol.Command) chan int {
	connectionState := make(chan int)
	go func() {
		for {
			quit := make(chan bool)
			log.Info("Connection to ", config.Host, ":", config.Port)
			connection, err := net.Dial("tcp", net.JoinHostPort(config.Host, config.Port))
			if err != nil {
				log.Error("Failed connection: ", err)
				<-time.After(time.Second)
				continue
			}

			connectionState <- ConnectionStateConnected
			log.Trace("Connected")
			go sendWorker(connection, send, quit)

			connectionWorker(connection, recv)
			close(quit)
			connectionState <- ConnectionStateDisconnected

			log.Warn("Lost connection, reconnecting")
			<-time.After(time.Second)
		}
	}()
	return connectionState
}

func sendWorker(connection net.Conn, send chan interface{}, quit chan bool) {
	var err error
	var ret []byte
	for {
		select {
		case d := <-send:

			if a, ok := d.(*protocol.Node); ok {
				a.Uuid = config.Uuid
				ret, err = json.Marshal(a)
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
			log.Info(cmd)
			recv <- cmd
		}

	}
}
