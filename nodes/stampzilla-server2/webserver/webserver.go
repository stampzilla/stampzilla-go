package webserver

import (
	"crypto/tls"
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
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/handlers"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/models"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/store"
)

type Webserver struct {
	Store            *store.Store
	Melody           *melody.Melody
	Config           *models.Config
	WebsocketHandler handlers.WebsocketHandler
	router           http.Handler
}

func New(s *store.Store, conf *models.Config, wsh handlers.WebsocketHandler, m *melody.Melody) *Webserver {

	return &Webserver{
		Store:            s,
		Config:           conf,
		WebsocketHandler: wsh,
		Melody:           m,
	}
}

func (ws *Webserver) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ws.router.ServeHTTP(w, req)
}

func (ws *Webserver) Init() *gin.Engine {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(gzip.Gzip(gzip.DefaultCompression))

	ws.initMelody()

	r.Use(ginlogrus.New(logrus.StandardLogger()))

	// Setup gin
	csp := r.Group("/")
	csp.Use(cspMiddleware())
	csp.StaticFile("/", "./web/dist/index.html")
	csp.StaticFile("/index.html", "./web/dist/index.html")
	csp.StaticFile("/service-worker.js", "./web/dist/service-worker.js")
	r.Static("/assets", "./web/dist/assets")
	r.GET("/ca.crt", ws.handleDownloadCA())

	r.GET("/ws", ws.handleWs(ws.Melody))

	ws.router = r
	return r
}
func (ws *Webserver) Start(addr string, tlsConfig *tls.Config) chan struct{} {

	server, done := gograce.NewServerWithTimeout(10 * time.Second)

	server.Handler = ws.Init()
	server.Addr = addr

	go func() {
		if tlsConfig != nil {
			server.TLSConfig = tlsConfig
			logrus.Infof("Starting secure webserver at %s", addr)
			logrus.Error(server.ListenAndServeTLS("", ""))
		} else {
			logrus.Infof("Starting webserver at %s", addr)
			logrus.Error(server.ListenAndServe())
		}
	}()
	return done
}

func (ws *Webserver) initMelody() {
	// Setup melody
	ws.Melody.Upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws.Melody.HandleConnect(ws.handleConnect(ws.Store))
	ws.Melody.HandleMessage(ws.handleMessage(ws.Store))
	ws.Melody.HandleDisconnect(ws.handleDisconnect(ws.Store))
}

func cspMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Security-Policy", "worker-src 'self';")
		c.Next()
	}
}

func (ws *Webserver) handleConnect(store *store.Store) func(s *melody.Session) {
	return func(s *melody.Session) {
		t, exists := s.Get("protocol")
		if !exists {
			logrus.Error("No Sec-WebSocket-Protocol defined. Aborting")
			return
		}

		id, _ := s.Get("ID")

		store.AddOrUpdateConnection(id.(string), &models.Connection{
			Type:       t.(string),
			RemoteAddr: s.Request.RemoteAddr,
			Attributes: s.Keys,
		})

		err := ws.WebsocketHandler.Connect(s, s.Request, s.Keys)
		if err != nil {
			logrus.Error(err)
			return
		}

	}
}

func (ws *Webserver) handleMessage(store *store.Store) func(s *melody.Session, msg []byte) {
	return func(s *melody.Session, msg []byte) {
		data, err := models.ParseMessage(msg)
		if err != nil {
			logrus.Error("cannot parse incoming websocket message: ", err)
			return
		}

		id, _ := s.Get("ID")
		data.FromUUID = id.(string)
		err = ws.WebsocketHandler.Message(data)
		if err != nil {
			logrus.Error(err)
			return
		}

	}
}

func (ws *Webserver) handleDisconnect(store *store.Store) func(s *melody.Session) {
	return func(s *melody.Session) {
		err := ws.WebsocketHandler.Disconnect(s)
		if err != nil {
			logrus.Error(err)
			return
		}
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
		//TODO check if node sent uuid then use that

		keys["ID"] = uuid.String()

		if c.Request.TLS != nil {
			keys["secure"] = true

			certs := c.Request.TLS.PeerCertificates
			if len(certs) > 0 {
				keys["identity"] = certs[0].Subject.CommonName
			}
		}

		// Accept the requested protocol
		// TODO: only accept known protocols
		if c.Request.Header.Get("Sec-WebSocket-Protocol") != "" {
			c.Writer.Header().Set("Sec-WebSocket-Protocol", c.Request.Header.Get("Sec-WebSocket-Protocol"))
			keys["protocol"] = c.Request.Header.Get("Sec-WebSocket-Protocol")
		}

		if c.Request.Header.Get("X-UUID") != "" {
			keys["ID"] = c.Request.Header.Get("X-UUID")
		}
		if c.Request.Header.Get("X-TYPE") != "" {
			keys["type"] = c.Request.Header.Get("X-TYPE")
		}

		if ws.Store.Connection(keys["ID"].(string)) != nil {
			logrus.Error("Connection with same UUID already exists")
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		m.HandleRequestWithKeys(c.Writer, c.Request, keys)
	}
}
