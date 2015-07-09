package main

import (
	"net/http"

	sockets "github.com/beatrichartz/martini-sockets"
	log "github.com/cihub/seelog"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/encoder"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/websocket"
)

// Webserver that serves static files

type WebServer struct {
	Config     *ServerConfig      `inject:""`
	WsClients  *websocket.Clients `inject:""`
	WebHandler *WebHandler        `inject:""`
}

func NewWebServer() *WebServer {
	return &WebServer{}
}

func (ws *WebServer) Start() {
	log.Info("Starting WEB (:" + ws.Config.WebPort + " in " + ws.Config.WebRoot + ")")

	//m := martini.Classic()
	r := martini.NewRouter()
	ma := martini.New()
	//ma.Use(martini.Logger())
	ma.Use(martini.Recovery())
	ma.Use(martini.Static(ws.Config.WebRoot, martini.StaticOptions{SkipLogging: true}))
	ma.MapTo(r, (*martini.Routes)(nil))
	ma.Action(r.Handle)
	m := &martini.ClassicMartini{ma, r}

	m.Use(func(c martini.Context, w http.ResponseWriter) {
		c.MapTo(encoder.JsonEncoder{}, (*encoder.Encoder)(nil))
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
	})

	m.Get("/socket", sockets.JSON(websocket.Message{}), ws.WsClients.WebsocketRoute)

	//Nodes
	m.Get("/api/nodes", ws.WebHandler.GetNodes)
	m.Get("/api/nodes/:id", ws.WebHandler.GetNode)
	m.Put("/api/nodes/:id/cmd", ws.WebHandler.CommandToNodePut)
	m.Get("/api/nodes/:id/cmd/**", ws.WebHandler.CommandToNodeGet)

	//Rules
	m.Get("/api/rules", ws.WebHandler.GetRules)

	//Schedule
	m.Get("/api/schedule", ws.WebHandler.GetScheduleTasks)
	m.Get("/api/schedule/entries", ws.WebHandler.GetScheduleEntries)

	m.Get("/api/reload", ws.WebHandler.GetReload)

	// Server state methods
	m.Get("/api/trigger/:key/:value", ws.WebHandler.GetServerTrigger)
	m.Get("/api/set/:key/:value", ws.WebHandler.GetServerSet)

	go func() {
		log.Critical(http.ListenAndServe(":"+ws.Config.WebPort, m))
	}()
}
