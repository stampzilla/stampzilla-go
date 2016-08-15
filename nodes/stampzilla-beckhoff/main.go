package main

import (
	"flag"
	"os"
	"os/signal"
	"strconv"
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
var settings *Config

var VERSION string = "dev"
var BUILD_DATE string = ""

func main() {
	config := basenode.NewConfig()
	basenode.SetConfig(config)

	settings = NewConfig()
	SetConfig(settings)

	flag.Parse()

	goADS.UseLogger(log.Current)
	log.Info("Starting beckhoff node")

	// Create new node description
	node = protocol.NewNode("beckhoff")
	node.Version = VERSION
	node.BuildDate = BUILD_DATE

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

	connection.Connect()

	if settings.Tpy != "" {
		symbols = connection.ParseTPY(settings.Tpy)
	} else {
		symbols, _ = connection.UploadSymbolInfo()

		log.Debug("Symbols uploaded:")
		for key, _ := range symbols {
			log.Debug(" - ", key)
		}
	}

	// Check what device are we connected to/*{{{*/
	data, e := connection.ReadDeviceInfo()
	if e != nil {
		log.Critical(e)
		os.Exit(1)
	}
	log.Infof("Successfully conncected to \"%s\" version %d.%d (build %d)", data.DeviceName, data.MajorVersion, data.MinorVersion, data.BuildVersion) /*}}}*/

	for variableName, variable := range settings.Variables {
		foundSymbol := Find(variableName)
		if foundSymbol != nil {
			foundSymbol.Read()
			WalkSymbol(variable, foundSymbol)

			variable.symbol = foundSymbol
			variable.symbolCtrl = Find(variable.Ctrl)

			foundSymbol.AddDeviceNotification(func(symbol *goADS.ADSSymbol) {
				WalkSymbol(variable, symbol)
				serverConnection.Send(node.Node())
			})

		}
	}

	serverConnection.Send(node.Node())

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

func Find(variableName string) *goADS.ADSSymbol {
	for _, symbol := range symbols {
		if len(variableName) < len(symbol.FullName) {
			continue
		}
		if symbol.FullName != variableName[:len(symbol.FullName)] {
			continue
		}

		found := symbol.Find(variableName)
		if len(found) > 0 {
			return found[0]
		}
	}

	return nil
}

/*
type Symbol goADS.ADSSymbol;
func (data *Symbol) Find(name string) *Symbol {
	/*if len(data.Childs) == 0 {

	} else {
		for i, _ := range data.Childs {
		}
	}

	return nil;
}
*/

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
		variable, ok := settings.Variables[name]
		if ok {
			var target string

			if len(cmd.Params) == 1 {
				target = cmd.Params[0]
				//WriteSymbol(&iface, name, cmd.Params[0])
			} else if len(cmd.Args) == 2 {
				target = cmd.Args[1]
				//WriteSymbol(&iface, name, cmd.Args[1])
			}

			if variable.Max != 0 || variable.Min != 0 {
				i, _ := strconv.Atoi(target)
				var tmp float64
				tmp = (float64(i)/100)*(variable.Max-variable.Min) + variable.Min
				i = int(tmp)
				target = strconv.Itoa(i)
			}

			if variable.symbol != nil {
				variable.symbol.Write(target)
			}

			if variable.symbolCtrl != nil {
				variable.symbolCtrl.Write(target)
			}
		} else {
			log.Warn("Tag ", name, " not found in beckhoff.json")
		}
	}

	return nil
}

func WalkSymbol(variable *Variable, data *goADS.ADSSymbol) { /*{{{*/
	if len(data.Childs) == 0 {

		name := strings.Replace(data.FullName, ".", "_", -1)
		val, ok := state.Values[name]

		if !ok {
			val = StateValue{symbol: data}
		}

		if !data.Valid {
			val.String = "INVALID"
		} else {
			val.Type = data.DataType
			val.String = data.Value
			val.Bool = data.Value == "True"

			switch data.DataType {
			case "UINT":
				val.Int, _ = strconv.Atoi(data.Value)
				if variable.Max != 0 || variable.Min != 0 {
					var tmp float64
					tmp = (float64(val.Int) - variable.Min) / (variable.Max - variable.Min) * 100
					val.Int = int(tmp)
					val.Bool = (float64(val.Int) - variable.Min) > 0
				}
			}
		}

		if !ok {
			switch variable.Type {
			case "slider":
				node.AddElement(&protocol.Element{
					Type: protocol.ElementTypeSlider,
					Name: variable.Name,
					Command: &protocol.Command{
						Cmd:  "set",
						Args: []string{name},
					},
					Feedback: "Values." + name + ".Int",
				})
			case "toggle":
				node.AddElement(&protocol.Element{
					Type: protocol.ElementTypeToggle,
					Name: variable.Name,
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
			WalkSymbol(variable, data.Childs[i].Self)
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
