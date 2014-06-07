package main

import (
	//"code.google.com/p/go.net/websocket"
	"encoding/json"
	"fmt"
	log "github.com/cihub/seelog"
	"github.com/go-martini/martini"
	"github.com/gorilla/websocket"
	//"github.com/martini-contrib/binding"
	"github.com/martini-contrib/encoder"
	"net/http"
)

// Webserver that serves static files
var chttp = http.NewServeMux()

var WebSockets map[int]chan string
var WebSocketPointer int

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
	//chttp.Handle("/", http.FileServer(http.Dir("./site/")))
	//http.HandleFunc("/", webHandler)
	//http.Handle("/socket", websocket.Handler(socketServer))

	WebSockets = make(map[int]chan string)
	m := martini.Classic()
	//m.Use(martini.Static("site"))

	m.Use(func(c martini.Context, w http.ResponseWriter) {
		c.MapTo(encoder.JsonEncoder{}, (*encoder.Encoder)(nil))
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
	})

	m.Get("/socket", socketServer)
	m.Get("/api/nodes", GetNodes)
	m.Get("/api/node/:id", GetNode)
	m.Post("/api/node/:id/state", PostNodeState)
	m.Get("/api/users/:id", GetUser)

	//handler.SetRoutes(
	//rest.Route{"GET", "/api/nodes", GetNodes},
	//rest.Route{"GET", "/api/node/:id", GetNode},
	//rest.Route{"POST", "/api/node/:id/state", PostNodeState},
	//rest.Route{"GET", "/api/users/:id", GetUser},
	//)

	//go http.ListenAndServe(":"+port, nil)
	log.Critical(http.ListenAndServe(":"+port, m))
	//fmt.Println()
} /*}}}*/
//func webHandler(w http.ResponseWriter, r *http.Request) { [>{{{<]
//// Select what to serve
//if len(r.URL.Path) > 3 && r.URL.Path[1:4] == "api" {
//handler.ServeHTTP(w, r)
//return
//}

//switch r.URL.Path[1:] {
//case "": // The main page (index.htm)
//// Handle simulate tag scans
//if r.FormValue("tag") != "" {
////
//}

//// Serve the index.htm
//body, _ := ioutil.ReadFile("./site/index.htm")
//fmt.Fprintf(w, string(body))
//default: // Serve all files in the web folder
//chttp.ServeHTTP(w, r)
//}
//} [>}}}<]

type T struct {
	Msg string
}

func socketServer(w http.ResponseWriter, r *http.Request) { /*{{{*/

	ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(w, "Not a websocket handshake", 400)
		return
	} else if err != nil {
		log.Error(err)
		return
	}
	fmt.Printf("jsonServer %#v\n")

	pointer := WebSocketPointer
	WebSocketPointer++

	WebSockets[pointer] = make(chan string)
	defer func() { WebSockets[pointer] = nil }()
	defer close(WebSockets[pointer])

	//var msg T

Main:
	for {
		select {
		//case <-time.After(time.Second / 1):
		//msg.Msg = "$(\".navbar-brand\").html(\"" + time.Now().Fomat(time.StampMilli) + "\")"
		//err := websocket.JSON.Send(ws, msg)

		////buf := []byte("test")
		////_, err := ws.Write(buf[:])
		//if err != nil {
		//fmt.Println(err)
		//break Main
		//}
		//fmt.Printf("send:%q\n", buf[:])
		case txt := <-WebSockets[pointer]:
			log.Critical("Worker recived data: ", txt)
			//websocket.Message.Send(ws, txt)
			err := ws.WriteMessage(1, []byte(txt))
			if err != nil {
				fmt.Println(err)
				break Main
			}
		}
		continue

		// Receive receives a text message serialized T as JSON.
		_, msg, err := ws.ReadMessage()
		if err != nil {
			fmt.Println(err)
			break
		}
		fmt.Printf("recv:%#v\n", msg)
		// Send send a text message serialized T as JSON.
		err = ws.WriteJSON([]byte(msg))
		//err = websocket.JSON.Send(ws, msg)
		if err != nil {
			fmt.Println(err)
			break
		}
		fmt.Printf("send:%#v\n", msg)
	}
} /*}}}*/

func GetUser(w *http.ResponseWriter, req *http.Request) {

}
