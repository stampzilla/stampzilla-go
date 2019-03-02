package webserver

import (
	"crypto/tls"
	"io"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jonaz/ginlogrus"
	"github.com/jonaz/gograce"
	"github.com/olahol/melody"
	"github.com/rakyll/statik/fs"
	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/handlers"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models"

	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/store"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/websocket"
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
	r.Use(cors.Default())

	statikFS, err := fs.New()
	if err == nil { // we only service GUI if statik files can be found
		r.GET("/service-worker.js", gin.WrapH(http.FileServer(statikFS)))
		r.GET("/assets/*all", gin.WrapH(http.FileServer(statikFS)))
		r.NoRoute(func(c *gin.Context) {
			cspMiddleware()(c)
			c.Request.URL.Path = "/" // force us to always return index.html and not the requested page to be compatible with HTML5 routing
			http.FileServer(statikFS).ServeHTTP(c.Writer, c.Request)
		})
	}

	// Setup gin
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
	ws.Melody.HandleConnect(ws.handleConnect())
	ws.Melody.HandleMessage(ws.handleMessage())
	ws.Melody.HandleDisconnect(ws.handleDisconnect())
}

func cspMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Security-Policy", "worker-src 'self';")
		c.Next()
	}
}

func (ws *Webserver) handleConnect() func(s *melody.Session) {
	return func(s *melody.Session) {
		proto, _ := s.Get(websocket.KeyProtocol.String())
		id, _ := s.Get(websocket.KeyID.String())

		ws.Store.AddOrUpdateConnection(id.(string), &models.Connection{
			Type:       proto.(string),
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

func (ws *Webserver) handleMessage() func(s *melody.Session, msg []byte) {
	return func(s *melody.Session, msg []byte) {
		data, err := models.ParseMessage(msg)
		if err != nil {
			logrus.Error("cannot parse incoming websocket message: ", err)
			return
		}

		id, _ := s.Get(websocket.KeyID.String())
		data.FromUUID = id.(string)
		err = ws.WebsocketHandler.Message(s, data)
		if err != nil {
			logrus.Error(err)
			return
		}
	}
}

func (ws *Webserver) handleDisconnect() func(s *melody.Session) {
	return func(s *melody.Session) {
		id, _ := s.Get(websocket.KeyID.String())
		ws.Store.RemoveConnection(id.(string))
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

		file, err := os.Open(path.Join("certificates", "ca.crt"))
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
		keys := make(map[string]interface{})
		keys[websocket.KeyID.String()] = uuid.New().String()

		if c.Request.TLS != nil {
			keys["secure"] = true

			certs := c.Request.TLS.PeerCertificates
			if len(certs) > 0 {
				keys["identity"] = certs[0].Subject.CommonName
			}
		}

		// Accept the requested protocol if known
		knownProtocols := []string{
			"node",
			"gui",
			"metrics",
		}
		proto := c.Request.Header.Get("Sec-WebSocket-Protocol")

		allowed := false
		for _, v := range knownProtocols {
			if proto == v {
				allowed = true
				break
			}
		}

		if !allowed {
			logrus.Errorf("webserver: protocol \"%s\" not allowed", proto)
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		if proto != "" {
			c.Writer.Header().Set("Sec-WebSocket-Protocol", proto)
			keys[websocket.KeyProtocol.String()] = proto
		}

		if c.Request.Header.Get("X-UUID") != "" {
			keys[websocket.KeyID.String()] = c.Request.Header.Get("X-UUID")
		}
		keys["type"] = c.Request.Header.Get("X-TYPE")

		if ws.Store.Connection(keys[websocket.KeyID.String()].(string)) != nil {
			logrus.Error("Connection with same UUID already exists")
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		err := m.HandleRequestWithKeys(c.Writer, c.Request, keys)
		if err != nil {
			logrus.Errorf("webserver: %s", err.Error())
			return
		}

	}
}
