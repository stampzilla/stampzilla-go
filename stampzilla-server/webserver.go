package main

import (
	"net/http"

	"github.com/beatrichartz/martini-sockets"
	log "github.com/cihub/seelog"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/encoder"
	"github.com/stampzilla/stampzilla-go/stampzilla-server/logic"
	serverprotocol "github.com/stampzilla/stampzilla-go/stampzilla-server/protocol"
)

// Webserver that serves static files

type WebServer struct {
	Config     *ServerConfig         `inject:""`
	Logic      *logic.Logic          `inject:""`
	Nodes      *serverprotocol.Nodes `inject:""`
	WsClients  *Clients              `inject:""`
	WebHandler *WebHandler           `inject:""`
}

func NewWebServer() *WebServer {
	return &WebServer{}
}

func (ws *WebServer) Start() {
	log.Info("Starting WEB (:" + ws.Config.WebPort + " in " + ws.Config.WebRoot + ")")

	//m := martini.Classic()
	r := martini.NewRouter()
	ma := martini.New()
	ma.Use(martini.Logger())
	ma.Use(martini.Recovery())
	ma.Use(martini.Static(ws.Config.WebRoot))
	ma.MapTo(r, (*martini.Routes)(nil))
	ma.Action(r.Handle)
	m := &martini.ClassicMartini{ma, r}

	m.Use(func(c martini.Context, w http.ResponseWriter) {
		c.MapTo(encoder.JsonEncoder{}, (*encoder.Encoder)(nil))
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
	})

	m.Get("/socket", sockets.JSON(Message{}), ws.WsClients.websocketRoute)

	//Nodes
	m.Get("/api/nodes", ws.WebHandler.GetNodes)
	m.Get("/api/node/:id", ws.WebHandler.GetNode)
	m.Put("/api/node/:id/cmd", ws.WebHandler.CommandToNode)

	//Rules
	m.Get("/api/rules", ws.WebHandler.GetRules)

	log.Critical(http.ListenAndServe(":"+ws.Config.WebPort, m))
}
