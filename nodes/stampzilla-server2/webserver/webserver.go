package webserver

import (
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jonaz/ginlogrus"
	"github.com/jonaz/gograce"
	"github.com/olahol/melody"
	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/models"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/store"
)

type Webserver struct {
	Store  *store.Store
	Melody *melody.Melody
	Config *models.Config
}

func New(s *store.Store, conf *models.Config) *Webserver {

	return &Webserver{
		Store:  s,
		Config: conf,
	}

}

func (ws *Webserver) Init() *gin.Engine {

	r := gin.New()
	r.Use(gzip.Gzip(gzip.DefaultCompression))

	m := ws.initMelody()

	ws.Melody = m

	r.Use(ginlogrus.New(logrus.StandardLogger()))

	// Setup gin
	csp := r.Group("/")
	csp.Use(cspMiddleware())
	csp.StaticFile("/", "./web/dist/index.html")
	csp.StaticFile("/index.html", "./web/dist/index.html")
	csp.StaticFile("/service-worker.js", "./web/dist/service-worker.js")
	r.Static("/assets", "./web/dist/assets")
	r.GET("/ca.crt", ws.handleDownloadCA())

	r.GET("/ws", ws.handleWs(m))

	return r
}
func (ws *Webserver) Start(addr string) chan struct{} {

	server, done := gograce.NewServerWithTimeout(10 * time.Second)

	server.Handler = ws.Init()
	server.Addr = addr

	go func() {
		logrus.Error(server.ListenAndServe())
	}()
	return done
}

func (ws *Webserver) initMelody() *melody.Melody {
	// Setup melody
	m := melody.New()
	m.Upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	m.HandleConnect(ws.handleConnect(ws.Store))
	m.HandleMessage(ws.handleMessage(m, ws.Store))
	m.HandleDisconnect(ws.handleDisconnect(ws.Store))
	return m
}

func cspMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Security-Policy", "worker-src 'self';")
		c.Next()
	}
}

func (ws *Webserver) handleConnect(store *store.Store) func(s *melody.Session) {
	return func(s *melody.Session) {
		id, _ := s.Get("ID")
		t, _ := s.Get("protocol")

		store.AddOrUpdateConnection(id.(string), &models.Connection{
			Type:       t.(string),
			RemoteAddr: s.Request.RemoteAddr,
			Attributes: s.Keys,
		})

		//if exists && t == "gui" {
		msg, err := models.NewMessage("server-info", models.ServerInfo{
			Name:    ws.Config.Name,
			UUID:    ws.Config.UUID,
			TLSPort: ws.Config.TLSPort,
			Port:    ws.Config.Port,
		})
		if err != nil {
			logrus.Error(err)
			return
		}
		msg.Write(s)
		//}
	}
}

func (ws *Webserver) handleMessage(m *melody.Melody, store *store.Store) func(s *melody.Session, msg []byte) {
	return func(s *melody.Session, msg []byte) {
		data, err := models.ParseMessage(msg)
		if err != nil {
			logrus.Error(err)
			return
		}

		logrus.Warn(data)

		//switch data.Type {
		//case "update-node":
		//ws.handleNodeUpdate(m, s, store, data)
		//}
	}
}

func (ws *Webserver) handleDisconnect(store *store.Store) func(s *melody.Session) {
	return func(s *melody.Session) {
		id, _ := s.Get("ID")
		store.RemoveConnection(id.(string))
	}
}

func (ws *Webserver) handleDownloadCA() func(c *gin.Context) {
	return func(c *gin.Context) {
		header := c.Writer.Header()
		header["Content-Type"] = []string{"application/x-x509-ca-cert"}

		file, err := os.Open("ca.crt")
		if err != nil {
			c.String(http.StatusOK, "%v", err)
			return
		}
		defer file.Close()

		io.Copy(c.Writer, file)
	}
}

func (ws *Webserver) handleWs(m *melody.Melody) func(c *gin.Context) {
	return func(c *gin.Context) {
		uuid := uuid.New()
		keys := make(map[string]interface{})
		keys["ID"] = uuid.String()

		if c.Request.TLS != nil {
			certs := c.Request.TLS.PeerCertificates
			logrus.Warn("HTTP CERTS", certs)
			keys["secure"] = true
		}

		// Accept the requested protocol
		// TODO: only accept known protocols
		if c.Request.Header.Get("Sec-WebSocket-Protocol") != "" {
			c.Writer.Header().Set("Sec-WebSocket-Protocol", c.Request.Header.Get("Sec-WebSocket-Protocol"))
			keys["protocol"] = c.Request.Header.Get("Sec-WebSocket-Protocol")
		}

		m.HandleRequestWithKeys(c.Writer, c.Request, keys)
	}
}
