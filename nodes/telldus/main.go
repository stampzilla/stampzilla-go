package main

import (
    "encoding/json"
    "flag"
    "fmt"
    log "github.com/cihub/seelog"
    "net"
    "os/exec"
    "regexp"
)

var Info InfoStruct
var devices []Device

var host string
var port string

type Device struct { /*{{{*/
    Id       string
    Name     string
    State    string
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
    Section string
}                   /*}}}*/
type State struct { /*{{{*/
    Devices []Device
}   /*}}}*/

func main() {
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

    b, err := json.Marshal(Info)
    if err != nil {
        log.Error(err)
    }

    log.Info("Connect to ", host, ":", port)
    c, e := net.Dial("tcp", net.JoinHostPort(host, port))
    if e != nil {
        log.Error(e)
    } else {
        fmt.Fprintf(c, string(b))
    }

    select {}
}

func updateActions() {
    Info.Actions = []Action{
        Action{
            "set",
            "Set",
            []string{"Devices.Id"},
        },
    }
}

func updateLayout() {
    Info.Layout = []Layout{
        Layout{
            "1",
            "switch",
            "toggle",
            "Devices[Type=!dimmable]",
            "Lamps",
        },
    }
}

func readState() {
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
    findDevices := regexp.MustCompile("(?m)^(.+)\t(.+)\t(.*)$")
    if result := findDevices.FindAllStringSubmatch(string(out), -1); len(result) > 0 {
        for _, dev := range result {
            devices = append(devices, Device{dev[1], dev[2], dev[3], []string{"toggle"}})
        }
    }

    Info.State.Devices = devices

    log.Debug(devices)
}
