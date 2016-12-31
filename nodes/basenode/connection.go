package basenode

import (
	"encoding/json"
	"net"
	"time"

	log "github.com/cihub/seelog"
	"github.com/davecgh/go-spew/spew"
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
	send                   chan interface{}
	receive                chan protocol.Command
	receiveDeviceConfigSet chan protocol.DeviceConfigSet
	state                  chan int
}

func (c *connection) ReceiveCommands() chan protocol.Command {
	return c.receive
}
func (c *connection) ReceiveDeviceConfigSet() chan protocol.DeviceConfigSet {
	return c.receiveDeviceConfigSet
}
func (c *connection) Receive() chan protocol.Command {
	return c.receive
}
func (c *connection) State() chan int {
	return c.state
}
func (c *connection) Send(data interface{}) {
	select {
	case c.send <- data:
	case <-time.After(time.Second * 10):
	}
}

func Connect() *connection {
	connection := &connection{
		send:                   make(chan interface{}, 100),
		receive:                make(chan protocol.Command, 100),
		receiveDeviceConfigSet: make(chan protocol.DeviceConfigSet, 100),
		state: make(chan int),
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

			connection.connectionWorker(tcpConnection, serverIsAlive)
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
			switch a := d.(type) {
			case *protocol.Node:
				a.SetUuid(config.Uuid)

				pkg := protocol.NewUpdateWithData(protocol.TypeUpdateNode, a.Node())
				//log.Trace("Sending node package: ", spew.Sdump(pkg))
				err = encoder.Encode(pkg)
			case notifications.Notification:
				pkg := protocol.NewUpdateWithData(protocol.TypeNotification, a)
				//log.Trace("Sending notification: ", spew.Sdump(pkg))
				err = encoder.Encode(pkg)
			default:
				//log.Tracef("Sending %T package: %#v", d, d)
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

func (conn *connection) connectionWorker(connection net.Conn, serverIsAlive chan bool) {
	// Recive data
	decoder := json.NewDecoder(connection)
	for {
		msg := protocol.NewUpdate()
		err := decoder.Decode(&msg)

		if err != nil {
			if err.Error() == "EOF" {
				connection.Close()
				log.Error("EOF:", err)
				return
			}
			connection.Close()
			log.Warn(err)
			return
		}

		serverIsAlive <- true

		switch msg.Type {
		case protocol.TypePing:
			//log.Debug("Recived ping - pong")
			//PONG packet
			connection.Write([]byte(`{"Type":5}`))
		case protocol.TypeCommand:
			var cmd protocol.Command
			err = json.Unmarshal(*msg.Data, &cmd)
			if err != nil {
				log.Debug("Failed to decode command from server", err)
				spew.Dump(string(*msg.Data))
				continue
			}

			log.Debug("Command from server", cmd)
			conn.receive <- cmd
		case protocol.TypeDeviceConfigSet:
			var cmd protocol.DeviceConfigSet
			err = json.Unmarshal(*msg.Data, &cmd)
			if err != nil {
				log.Debug("Failed to decode DeviceConfigSet from server", err)
				spew.Dump(string(*msg.Data))
				continue
			}

			log.Debug("DeviceConfigSet from server", cmd)
			conn.receiveDeviceConfigSet <- cmd
		default:
			log.Warnf("Received a %s from server, dont know what to do... throwing it away", msg.Type)
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
			//connection.Write([]byte("{\"Ping\":true}"))
			//PING packet
			connection.Write([]byte(`{"Type":4}`))

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
