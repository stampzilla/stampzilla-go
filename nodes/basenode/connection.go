package basenode

import (
	"encoding/json"
	"net"
	"time"

	log "github.com/cihub/seelog"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/notifications"
	"github.com/stampzilla/stampzilla-go/protocol"
)

const (
	ConnectionStateConnected    = 1
	ConnectionStateDisconnected = 0
)

type Sendable interface {
	Send(interface{})
}

type Connection interface {
	Sendable
	Receive() chan protocol.Command
	State() chan int
}

type connection struct {
	send    chan interface{}
	receive chan protocol.Command
	state   chan int
}

func (c *connection) Receive() chan protocol.Command {
	return c.receive
}
func (c *connection) State() chan int {
	return c.state
}
func (c *connection) Send(data interface{}) {
	c.send <- data
}

func Connect() Connection {

	connection := &connection{
		send:    make(chan interface{}, 100),
		receive: make(chan protocol.Command, 100),
		state:   make(chan int),
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

			connection.State() <- ConnectionStateConnected
			log.Trace("Connected")

			serverIsAlive := make(chan bool)
			go timeoutMonitor(tcpConnection, serverIsAlive)
			go sendWorker(tcpConnection, connection.send, quit)

			connectionWorker(tcpConnection, connection.receive, serverIsAlive)
			close(quit)
			connection.State() <- ConnectionStateDisconnected

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
				log.Trace("Sending node package: ", a)
				err = encoder.Encode(a.Node())
			} else if a, ok := d.(notifications.Notification); ok {
				type NotificationPkg struct {
					Notification notifications.Notification
				}
				note := NotificationPkg{
					Notification: a,
				}
				log.Tracef("Sending notification: %#v", d, d)
				err = encoder.Encode(note)
			} else {
				log.Tracef("Sending %T package: %#v", d, d)
				err = encoder.Encode(d)
			}
			if err != nil {
				log.Warn("Error encoder.Encode: ", err)
				connection.Close()
				log.Trace("sendWorker disconnected")
				return
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
				//log.Debug("Recived ping - pong")
				connection.Write([]byte("{\"Pong\":true}"))
				continue
			}

			log.Debug("Command from server", cmd)
			recv <- cmd
		}

	}
}

func timeoutMonitor(connection net.Conn, serverIsAlive chan bool) {
	log.Debug("Timeout monitor started (", connection.RemoteAddr(), ")")
	defer log.Debug("Timeout monitor closed (", connection.RemoteAddr(), ")")

	for {
		select {
		case <-serverIsAlive:
			// Everything is great, just continue
			continue
		case <-time.After(time.Second * 10):
			connection.Write([]byte("{\"Ping\":true}"))

			select {
			case <-serverIsAlive:
				continue
			case <-time.After(time.Second * 2):
				log.Warn("Server connection timeout, no answer to ping")
				connection.Close()
				return
			}
		}
	}
}
