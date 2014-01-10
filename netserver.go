package main

import (
    "fmt"
    "net"
)

func netStart() {
    l, err := net.Listen("tcp", ":8282")
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
    for {
        buf := make([]byte, 512)
        nr, err := c.Read(buf)
        if err != nil {
            return
        }

        data := buf[0:nr]
        println("Server got:", string(data))
        _, err = c.Write(data)
        if err != nil {
            fmt.Println("Failed write: ", err)
        }
    }
}
