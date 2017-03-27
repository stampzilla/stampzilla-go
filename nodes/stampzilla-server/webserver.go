//go:generate rm statik/statik.go
//go:generate go get github.com/rakyll/statik
//go:generate statik -src=./public/dist
package main

import (
	"net/http"

	log "github.com/cihub/seelog"
	"github.com/gin-gonic/gin"
	"github.com/rakyll/statik/fs"
	_ "github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/statik"
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

	//r := martini.NewRouter()
	//ma := martini.New()
	//ma.Use(martini.Recovery())
	//ma.Use(martini.Static(ws.Config.WebRoot, martini.StaticOptions{SkipLogging: true}))
	//ma.MapTo(r, (*martini.Routes)(nil))
	//ma.Action(r.Handle)
	//m := &martini.ClassicMartini{ma, r}

	//m.Use(func(c martini.Context, w http.ResponseWriter) {
	//c.MapTo(encoder.JsonEncoder{}, (*encoder.Encoder)(nil))
	//w.Header().Set("Content-Type", "application/json; charset=utf-8")
	//})

	r := gin.Default()

	statikFS, _ := fs.New()
	r.StaticFS("/app", statikFS)
	r.GET("/", func(c *gin.Context) { c.Redirect(302, "/app") })

	//m.Get("/socket", sockets.JSON(websocket.Message{}, &sockets.Options{AllowedOrigin: "https?://(localhost:5000|{{host}})$"}), ws.WsClients.WebsocketRoute)
	r.GET("/socket", ws.WsClients.WebsocketRoute)

	//Nodes
	r.GET("/api/nodes", ws.WebHandler.GetNodes)
	r.GET("/api/nodes/:id", ws.WebHandler.GetNode)
	r.PUT("/api/nodes/:id/cmd", ws.WebHandler.CommandToNodePut)
	r.GET("/api/nodes/:id/cmd/*cmd", ws.WebHandler.CommandToNodeGet)

	//Actions
	r.GET("/api/actions", ws.WebHandler.GetActions)
	r.GET("/api/actions/:uuid/run", ws.WebHandler.RunAction)
	r.GET("/api/actions/:uuid", ws.WebHandler.GetAction)

	//Rules
	r.GET("/api/rules", ws.WebHandler.GetRules)
	r.GET("/api/rules/:id/:action", ws.WebHandler.GetRunRules)

	//Schedule
	r.GET("/api/schedule", ws.WebHandler.GetScheduleTasks)
	r.GET("/api/schedule/entries", ws.WebHandler.GetScheduleEntries)

	// Server state methods
	r.GET("/api/trigger/:key/:value", ws.WebHandler.GetServerTrigger)
	r.GET("/api/set/:key/:value", ws.WebHandler.GetServerSet)

	//Devices
	r.GET("/api/devices", ws.WebHandler.GetDevices)

	go func() {
		log.Critical(http.ListenAndServe(":"+ws.Config.WebPort, r))
	}()
}
