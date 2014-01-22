package main

import (
	"encoding/json"
	"flag"
	"fmt"
	log "github.com/cihub/seelog"
	"io/ioutil"
	"net"
	"os/exec"
	"regexp"
	"strconv"
	"time"
)

var Info InfoStruct
var devices []Device

var Connection net.Conn

var host string
var port string

type Device struct { /*{{{*/
	Id       string
	Name     string
	State    string
	Type     string
	Features []string
}                        /*}}}*/
type InfoStruct struct { /*{{{*/
	Id      string
	Actions []Action
	Layout  []Layout
	State   State
}                    /*}}}*/
type Action struct { /*{{{*/
	Id        string
	Name      string
	Arguments []string
}                    /*}}}*/
type Layout struct { /*{{{*/
	Id      string
	Type    string
	Action  string
	Using   string
	Filter  []string
	Section string
}                   /*}}}*/
type State struct { /*{{{*/
	Devices []Device
} /*}}}*/

type Command struct {
	Cmd  string
	Args []string
}

func main() { /*{{{*/
	// Load logger
	logger, err := log.LoggerFromConfigAsFile("../logconfig.xml")
	if err != nil {
		panic(err)
	}
	log.ReplaceLogger(logger)

	// Load flags
	flag.StringVar(&host, "host", "localhost", "Stampzilla server hostname")
	flag.StringVar(&port, "port", "8282", "Stampzilla server port")
	flag.Parse()

	log.Info("Starting TELLDUS node")

	Info = InfoStruct{}
	Info.Id = "Tellstick"

	updateActions()
	updateLayout()
	readState()

	// Start the connection
	go connection()

	select {}
} /*}}}*/

func connection() {
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

		connectionWorker()

		log.Warn("Lost connection, reconnecting")
		<-time.After(time.Second)
	}
}

func connectionWorker() {
	// Send update
	sendUpdate()

	// Recive data
	for {
		buf := make([]byte, 51200)
		nr, err := Connection.Read(buf)
		if err != nil {
			return
		}

		data := buf[0:nr]

		var cmd Command
		err = json.Unmarshal(data, &cmd)
		if err != nil {
			log.Warn(err)
		} else {
			log.Info(cmd)
			processCommand(cmd)
		}

	}
}

func updateActions() { /*{{{*/
	Info.Actions = []Action{
		Action{
			"set",
			"Set",
			[]string{"Devices.Id"},
		},
		Action{
			"toggle",
			"Toggle",
			[]string{"Devices.Id"},
		},
		Action{
			"dim",
			"Dim",
			[]string{"Devices.Id", "value"},
		},
	}
}                     /*}}}*/
func updateLayout() { /*{{{*/
	Info.Layout = []Layout{
		Layout{
			"1",
			"switch",
			"toggle",
			//"Devices[Type=!dimmable]",
			"Devices",
			[]string{"check"},
			"Lamps",
		},
		Layout{
			"2",
			"slider",
			"dim",
			//"Devices[Type=dimmable]",
			"Devices",
			[]string{"dim"},
			"Lamps",
		},
	}
} /*}}}*/

func readState() { /*{{{*/
	out, err := exec.Command("tdtool", "--list").Output()
	if err != nil {
		log.Critical(err)
	}

	// Read number of devices
	cnt := regexp.MustCompile("Number of devices: ([0-9]+)?")
	if n := cnt.FindStringSubmatch(string(out)); len(n) > 1 {
		log.Debug("tdtool says ", n[1], " devices")
	}

	// Read all devices
	findDevices := regexp.MustCompile("(?m)^(.+)\t(.+)\t([A-Z]+|([A-Z]+):([0-9]+))$")
	if result := findDevices.FindAllStringSubmatch(string(out), -1); len(result) > 0 {
		for _, dev := range result {
			// fmt.Println(dev[1:])
			switch {
			case dev[4] == "DIMMED":
				devices = append(devices, Device{dev[1], dev[2], dev[5], "", []string{"check"}})
			case dev[3] == "OFF":
				devices = append(devices, Device{dev[1], dev[2], "false", "", []string{"check"}})
			case dev[3] == "ON":
				devices = append(devices, Device{dev[1], dev[2], "true", "", []string{"check"}})
			default:
				devices = append(devices, Device{dev[1], dev[2], dev[3], "", []string{"check"}})
			}
		}
	}

	Info.State.Devices = devices

	// Read all features from config
	config, _ := ioutil.ReadFile("/etc/tellstick.conf")
	findDevices = regexp.MustCompile("(?msU)device {.*id = ([0-9]+)[^0-9].*model = \"(.*)\".*^}$")
	if result := findDevices.FindAllStringSubmatch(string(config), -1); len(result) > 0 {
		for _, row := range result {
			for id, dev := range devices {
				if dev.Id == row[1] {
					devices[id].Type = row[2]

					switch row[2] {
					case "selflearning-dimmer":
						log.Warn("Found DIM at row ", id, " - ", dev, " | ", row)
						devices[id].Features = append(devices[id].Features, "dim")
					}
				}
			}
			//devices = append(devices, Device{dev[1], dev[2], dev[3], []string{"toggle"}})
		}
	}
	/*
		device {
		  id = 7
		  name = "tak bel."
		  protocol = "arctech"
		  model = "selflearning-dimmer"
		  parameters {
			house = "954"
			unit = "2"
		  }
		}
	*/

	//log.Debug(devices)
} /*}}}*/
func sendUpdate() {
	b, err := json.Marshal(Info)
	if err != nil {
		log.Error(err)
	}
	fmt.Fprintf(Connection, string(b))
}
func processCommand(cmd Command) {
	switch cmd.Cmd {
	case "toggle":
		for n, row := range Info.State.Devices {
			if row.Id == cmd.Args[0] {
				var arg = ""
				if Info.State.Devices[n].State == "false" {
					Info.State.Devices[n].State = "true"
					arg = "--on"
				} else {
					Info.State.Devices[n].State = "false"
					arg = "--off"
				}

				// Run command
				out, err := exec.Command("tdtool", arg, row.Id).Output()
				if err != nil {
					log.Critical(err)
				}
				log.Warn(string(out))

				sendUpdate()
				log.Info("Toggled=", Info.State.Devices[n].State)
			}
		}
	case "dim":
		for n, row := range Info.State.Devices {
			if row.Id == cmd.Args[0] {
				var dimmlevel float64
				val, _ := strconv.ParseFloat(cmd.Args[1], 64)
				dimmlevel = val * 255 / 100
				c := exec.Command("tdtool", "--dimlevel", fmt.Sprintf("%.0f", dimmlevel), "--dim", row.Id)
				out, err := c.Output()
				if err != nil {
					log.Critical(err)
				}
				log.Warn(c)
				log.Warn(string(out))

				Info.State.Devices[n].State = cmd.Args[1]

				sendUpdate()
				log.Info("Dimmed=", Info.State.Devices[n].State)
			}
		}

	}
}
