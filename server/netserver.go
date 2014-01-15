package main

import (
    "encoding/json"
    "fmt"
    log "github.com/cihub/seelog"
    "net"
)

var NodesConnection map[string]net.Conn
var NodesWait map[string]chan bool

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
}   /*}}}*/

func netStart(port string) {
    l, err := net.Listen("tcp", ":"+port)
    if err != nil {
        fmt.Println("listen error", err)
        return
    }

    NodesConnection = make(map[string]net.Conn)
    NodesWait = make(map[string]chan bool)

    go func() {
        for {
            fd, err := l.Accept()
            if err != nil {
                fmt.Println("accept error", err)
                return
            }

            go newClient(fd)
        }
    }()
}

// Handle a client
func newClient(c net.Conn) {
    log.Info("New client connected")
    id := ""
    for {
        buf := make([]byte, 51200)
        nr, err := c.Read(buf)
        if err != nil {
            log.Info(id, " - Client disconnected")
            if id != "" {
                delete(Nodes, id)
            }
            return
        }

        data := buf[0:nr]

        var info InfoStruct
        err = json.Unmarshal(data, &info)
        if err != nil {
            log.Warn(err)
        } else {
            id = info.Id
            Nodes[info.Id] = info
            NodesConnection[info.Id] = c

            log.Info(info.Id, " - Got update on state")

            if NodesWait[info.Id] != nil {
                select {
                case NodesWait[info.Id] <- false:
                    close(NodesWait[info.Id])
                    NodesWait[info.Id] = nil
                default:
                }
            }

            // Skicka till alla
            for n, _ := range WebSockets {
                if WebSockets[n] != nil {
                    select {
                    case WebSockets[n] <- string(data):
                    default:
                    }
                }
            }
        }

        /*_, err = c.Write(data)
          if err != nil {
              fmt.Println("Failed write: ", err)
          }*/
    }
}
