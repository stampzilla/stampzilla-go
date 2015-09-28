package main

import (
	"flag"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/stamp/goADS"

	log "github.com/cihub/seelog"
	"github.com/stampzilla/stampzilla-go/nodes/basenode"
	"github.com/stampzilla/stampzilla-go/protocol"
)

var WaitGroup sync.WaitGroup
var node *protocol.Node
var state *State = &State{}
var serverConnection basenode.Connection
var symbols map[string]goADS.ADSSymbol

func main() {
	config := basenode.NewConfig()
	basenode.SetConfig(config)

	settings := NewConfig()
	SetConfig(settings)

	flag.Parse()

	goADS.UseLogger(log.Current)
	log.Info("Starting Aquarium node")

	// Create new node description
	node = protocol.NewNode("beckhoff")
	state.Values = make(map[string]StateValue, 0)
	node.SetState(state)

	serverConnection = basenode.Connect()
	go monitorState(serverConnection)

	// This worker recives all incomming commands
	go serverRecv(serverConnection)

	// Startup the connection/*{{{*/
	connection, e := goADS.NewConnection(settings.Ip, settings.Netid, settings.Port)
	defer connection.Close() // Close the connection when we are done
	if e != nil {
		log.Critical(e)
		os.Exit(1)
	} /*}}}*/

	go shutdownRoutine(connection)

	if settings.Tpy != "" {
		symbols = connection.ParseTPY(settings.Tpy)
	} else {
		symbols, _ = connection.UploadSymbolInfo()
	}

	connection.Connect()

	// Check what device are we connected to/*{{{*/
	data, e := connection.ReadDeviceInfo()
	if e != nil {
		log.Critical(e)
		os.Exit(1)
	}
	log.Infof("Successfully conncected to \"%s\" version %d.%d (build %d)", data.DeviceName, data.MajorVersion, data.MinorVersion, data.BuildVersion) /*}}}*/

	iface, ok := symbols[".Interface"]
	if ok {
		iface.AddDeviceNotification(func(symbol *goADS.ADSSymbol) {
			WalkSymbol(symbol)
			serverConnection.Send(node.Node())
		})
	}

	go func() {
		for {
			select {
			case <-time.After(time.Second * 5):
				go connection.ReadState()
			}
		}
	}()

	WaitGroup.Wait()
	connection.Wait()
}

// WORKER that monitors the current connection state
func monitorState(connection basenode.Connection) {
	for s := range connection.State() {
		switch s {
		case basenode.ConnectionStateConnected:
			connection.Send(node.Node())
		case basenode.ConnectionStateDisconnected:
		}
	}
}

// WORKER that recives all incomming commands
func serverRecv(connection basenode.Connection) {
	for d := range connection.Receive() {
		if err := processCommand(d); err != nil {
			log.Error(err)
		}
	}
}

func processCommand(cmd protocol.Command) error {
	switch cmd.Cmd {
	case "set":
		name := strings.Replace(cmd.Args[0], "_", ".", -1)
		iface, ok := symbols[".Interface"]
		if ok {
			log.Info("Found .interface")
			if len(cmd.Params) == 1 {
				WriteSymbol(&iface, name, cmd.Params[0])
			} else if len(cmd.Args) == 2 {
				WriteSymbol(&iface, name, cmd.Args[1])
			}
		} else {
			log.Critical("Tag .interface not found")
		}
	}

	return nil
}

func WalkSymbol(data *goADS.ADSSymbol) { /*{{{*/
	if len(data.Childs) == 0 {
		name := strings.Replace(data.FullName, ".", "_", -1)
		val, ok := state.Values[name]

		if !ok {
			val = StateValue{}
		}

		if !data.Valid {
			val.String = "INVALID"
		} else {
			val.Type = data.DataType
			val.String = data.Value
			val.Bool = data.Value == "True"

			if !ok {
				node.AddElement(&protocol.Element{
					Type: protocol.ElementTypeToggle,
					Name: data.FullName,
					Command: &protocol.Command{
						Cmd:  "set",
						Args: []string{name},
					},
					Feedback: "Values." + name + ".Bool",
				})
			}
		}

		state.Values[name] = val
	} else {
		//log.Error("TYPE (", data.Area, ":", data.Offset, "): ", path, " [", data.DataType, "] = ", data.Value)
		for i, _ := range data.Childs {
			WalkSymbol(data.Childs[i].Self)
		}
	}
}                                                            /*}}}*/
func WriteSymbol(data *goADS.ADSSymbol, tag, value string) { /*{{{*/
	if len(data.Childs) == 0 {
		if data.FullName == tag {
			log.Error("Write to tag ", data.FullName)
			data.Write(value)
		} else {
			log.Debug("Not correct tag: ", tag, "!=", data.FullName)
		}
	} else {
		for i, _ := range data.Childs {
			WriteSymbol(data.Childs[i].Self, tag, value)
		}
	}
}                                              /*}}}*/
func shutdownRoutine(conn *goADS.Connection) { /*{{{*/
	sigchan := make(chan os.Signal, 2)
	signal.Notify(sigchan, os.Interrupt)
	signal.Notify(sigchan, syscall.SIGTERM)
	<-sigchan

	conn.Close()
} /*}}}*/
