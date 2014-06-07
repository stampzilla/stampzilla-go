package main

import (
    "encoding/json"
    "flag"
    "fmt"
    log "github.com/cihub/seelog"
    "net"
    "time"
)

// GLOBAL VARS
var Info InfoStruct

var Connection net.Conn

var host string
var port string

// TYPES
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
}                     /*}}}*/
type Command struct { /*{{{*/
    Cmd  string
    Args []string
}   /*}}}*/

// MAIN
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

    log.Info("Starting SIMPLE node")

    Info = InfoStruct{}
    Info.Id = "Simple"

    updateActions()
    updateLayout()
    updateState()

    // Start the connection, this gorutine will keep the connection up and reconnect if nessesary
    go connection()

    select {}
}   /*}}}*/

// CONNECTION
func connection() { /*{{{*/
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
}                         /*}}}*/
func connectionWorker() { /*{{{*/
    // When connected, send a description about this node
    sendUpdate()

    // Recive data
    for {
        // Create a buffer to save recive data to
        buf := make([]byte, 51200)

        // Wait for data and save it to buff. n = number of recived bytes
        nr, err := Connection.Read(buf)
        if err != nil {
            return
        }

        // Copy the recived bytes from buffer to data, length of data will be the same length as the recived bytes
        data := buf[0:nr]

        // Decode json command message
        var cmd Command
        err = json.Unmarshal(data, &cmd)
        if err != nil {
            log.Warn(err)
        } else {
            log.Info(cmd)

            // If decode is successfull, run command
            processCommand(cmd)
        }

    }
}                   /*}}}*/
func sendUpdate() { /*{{{*/
    b, err := json.Marshal(Info)
    if err != nil {
        log.Error(err)
    }
    fmt.Fprintf(Connection, string(b))
}   /*}}}*/

// INFORMATION
func updateActions() { /*{{{*/
    Info.Actions = []Action{
        Action{
            "toggle",
            "Toggle",
            []string{"Devices.Id"},
        },
    }
}                     /*}}}*/
func updateLayout() { /*{{{*/
    Info.Layout = []Layout{
        Layout{
            "1",
            "switch",
            "toggle",
            "Devices",
            []string{"check"},
            "Lamps",
        },
    }
}                    /*}}}*/
func updateState() { /*{{{*/
    Info.State = State{
        []Device{
            Device{
                Id:       "test",
                Name:     "asdasd",
                State:    "false",
                Type:     "toggle",
                Features: []string{"check"},
            },
        },
    }
}   /*}}}*/

// ACTIONS
func processCommand(cmd Command) {
    switch cmd.Cmd {
    case "toggle":
        for n, row := range Info.State.Devices {
            if row.Id == cmd.Args[0] {
                if Info.State.Devices[n].State == "false" {
                    Info.State.Devices[n].State = "true"
                } else {
                    Info.State.Devices[n].State = "false"
                }

                sendUpdate()
                log.Info("Toggled=", Info.State.Devices[n].State)
            }
        }
    }
}
