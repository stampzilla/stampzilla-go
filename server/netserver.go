package main

import (
    "encoding/json"
    "fmt"
    log "github.com/cihub/seelog"
    "net"
)

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

func netStart(port string) {
    l, err := net.Listen("tcp", ":"+port)
    if err != nil {
        fmt.Println("listen error", err)
        return
    }

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
    id := ""
    for {
        buf := make([]byte, 51200)
        nr, err := c.Read(buf)
        if err != nil {
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
        }

        /*_, err = c.Write(data)
          if err != nil {
              fmt.Println("Failed write: ", err)
          }*/
    }
}
