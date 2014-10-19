package main

import (
	"net/http"

	"github.com/beatrichartz/martini-sockets"
	log "github.com/cihub/seelog"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/encoder"
)

// Webserver that serves static files

func webStart(port, root string) {

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

	m.Get("/socket", sockets.JSON(Message{}), clients.websocketRoute)
	m.Get("/api/nodes", WebHandlerGetNodes)
	m.Get("/api/node/:id", WebHandlerGetNode)
	m.Put("/api/node/:id/cmd", WebHandlerCommandToNode)

	//Rules
	m.Get("/api/rules", WebHandlerGetRules)
	//m.Post("/api/node/:id/state", PostNodeState)
	//m.Get("/api/users/:id", GetUser)

	//go http.ListenAndServe(":"+port, nil)
	log.Critical(http.ListenAndServe(":"+port, m))
}
