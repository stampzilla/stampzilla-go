package main

import (
	"encoding/json"
	"github.com/beatrichartz/martini-sockets"
	log "github.com/cihub/seelog"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/encoder"
	"net/http"
)

// Webserver that serves static files

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
}                                  /*}}}*/
func webStart(port, root string) { /*{{{*/

	//m := martini.Classic()
	r := martini.NewRouter()
	ma := martini.New()
	ma.Use(martini.Logger())
	ma.Use(martini.Recovery())
	ma.Use(martini.Static(root))
	ma.MapTo(r, (*martini.Routes)(nil))
	ma.Action(r.Handle)
	m := &martini.ClassicMartini{ma, r}

	m.Use(func(c martini.Context, w http.ResponseWriter) {
		c.MapTo(encoder.JsonEncoder{}, (*encoder.Encoder)(nil))
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
	})

	m.Get("/socket", sockets.JSON(Message{}), websocketRoute)
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

func GetUser(w *http.ResponseWriter, req *http.Request) {

}
