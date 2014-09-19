package main

import (
	"flag"
	log "github.com/cihub/seelog"
	"github.com/stampzilla/stampzilla-go/protocol"
	"github.com/tarm/goserial"
	"io"
	"bytes"
	"encoding/binary"
	"strconv"
)

var node *protocol.Node
var c0 *SerialConnection;

var targetColor [4]byte;

type SerialConnection struct {
    Name string
    Baud int
	Port io.ReadWriteCloser
}

func main() {
	// Load logger
	logger, err := log.LoggerFromConfigAsFile("../logconfig.xml")
	if err != nil {
		panic(err)
	}
	log.ReplaceLogger(logger)

	// Load flags
	var host string
	var port string
	var dev string
	flag.StringVar(&host, "host", "localhost", "Stampzilla server hostname")
	flag.StringVar(&port, "port", "8282", "Stampzilla server port")
	flag.StringVar(&dev, "dev", "/dev/ttyACM0", "Arduino serial port")
	flag.Parse()

	// Create new node description
	node = protocol.NewNode("stamp-amber-lights")

	// Describe available actions
	node.AddAction("set", "Set", []string{"Devices.Id"})
	node.AddAction("toggle", "Toggle", []string{"Devices.Id"})
	node.AddAction("dim", "Dim", []string{"Devices.Id", "value"})

	// Describe available layouts
	node.AddLayout("1", "switch", "toggle", "Devices", []string{"on"}, "")
	node.AddLayout("2", "slider", "dim", "Devices", []string{"dim"}, "")
	node.AddLayout("3", "color-picker", "dim", "Devices", []string{"color"}, "")

	node.AddDevice("0","Color",[]string{"color"},"0");
	node.AddDevice("1","Red",[]string{"dim"},"0");
	node.AddDevice("2","Green",[]string{"dim"},"0");
	node.AddDevice("3","Blue",[]string{"dim"},"0");

	// Start the connection
	go connection(host, port, node)


	c0 = &SerialConnection{Name: dev, Baud: 9600}
	c0.connect();

	select {}
}

func processCommand(cmd protocol.Command) {

	type Cmd struct {
		Cmd uint16
		Arg uint32
	}

	type CmdColor struct {
		Cmd uint16
		Arg [4]byte
	}

	buf := new(bytes.Buffer)

	log.Info(cmd);

	switch cmd.Cmd {
	case "dim":
		value,_ := strconv.ParseInt(cmd.Args[1], 10, 32);

		value *= 255;
		value /= 100;

		switch(cmd.Args[0]) {
			case "1":
				targetColor[0] = byte(value);
			case "2":
				targetColor[1] = byte(value);
			case "3":
				targetColor[2] = byte(value);
		}

		err := binary.Write(buf, binary.BigEndian, &CmdColor{Cmd: 1, Arg: targetColor })
		if err != nil {
			log.Error("binary.Write failed:", err)
		}
	default:
		return;
	}

		n, err := c0.Port.Write(buf.Bytes())
		if err != nil {
			log.Error(err)
		}
		log.Info("Wrote ",n," bytes");
}


func (config *SerialConnection) connect() {

	c := &serial.Config{Name: config.Name, Baud: config.Baud}
	var err error

    config.Port, err = serial.OpenPort(c)
    if err != nil {
		log.Critical(err)
    }

	go func() {
		for {
			  buf := make([]byte, 128)

			  n, err := config.Port.Read(buf)
			  if err != nil {
					  log.Critical(err)
					return
			  }
			  log.Info("IN: ", string(buf[:n]) )
		}
	}()
}
