package main

import (
    "code.google.com/p/go.net/websocket"
    "encoding/json"
    "fmt"
    "github.com/stamp/go-json-rest"
    "io/ioutil"
    "net/http"
    "time"
)

// Webserver that serves static files
var chttp = http.NewServeMux()
var handler = rest.ResourceHandler{}

// The webserver
type Response map[string]interface{} /*{{{*/
func (r Response) String() (s string) {
    b, err := json.Marshal(r)
    if err != nil {
        s = ""
        return
    }
    s = string(b)
    return
}                            /*}}}*/
func webStart(port string) { /*{{{*/
    chttp.Handle("/", http.FileServer(http.Dir("./site/")))
    http.HandleFunc("/", webHandler)
    http.Handle("/socket", websocket.Handler(socketServer))

    handler.SetRoutes(
        rest.Route{"GET", "/api/nodes", GetNodes},
        rest.Route{"GET", "/api/node/:id", GetNode},
        rest.Route{"GET", "/api/users/:id", GetUser},
    )

    go http.ListenAndServe(":"+port, nil)
    fmt.Println()
}                                                         /*}}}*/
func webHandler(w http.ResponseWriter, r *http.Request) { /*{{{*/
    // Select what to serve
    if len(r.URL.Path) > 3 && r.URL.Path[1:4] == "api" {
        handler.ServeHTTP(w, r)
        return
    }

    switch r.URL.Path[1:] {
    case "": // The main page (index.htm)
        // Handle simulate tag scans
        if r.FormValue("tag") != "" {
            //
        }

        // Serve the index.htm
        body, _ := ioutil.ReadFile("./site/index.htm")
        fmt.Fprintf(w, string(body))
    default: // Serve all files in the web folder
        chttp.ServeHTTP(w, r)
    }
}   /*}}}*/

type T struct {
    Msg string
}

func socketServer(ws *websocket.Conn) {
    fmt.Printf("jsonServer %#v\n", ws.Config())

    var msg T

Main:
    for {
        select {
        case <-time.After(time.Second / 1):
            msg.Msg = "$(\".navbar-brand\").html(\"" + time.Now().Format(time.StampMilli) + "\")"
            err := websocket.JSON.Send(ws, msg)

            //buf := []byte("test")
            //_, err := ws.Write(buf[:])
            if err != nil {
                fmt.Println(err)
                break Main
            }
            //fmt.Printf("send:%q\n", buf[:])
        }
        continue

        // Receive receives a text message serialized T as JSON.
        err := websocket.JSON.Receive(ws, &msg)
        if err != nil {
            fmt.Println(err)
            break
        }
        fmt.Printf("recv:%#v\n", msg)
        // Send send a text message serialized T as JSON.
        err = websocket.JSON.Send(ws, msg)
        if err != nil {
            fmt.Println(err)
            break
        }
        fmt.Printf("send:%#v\n", msg)
    }
}

func GetUser(w *rest.ResponseWriter, req *rest.Request) {

}
