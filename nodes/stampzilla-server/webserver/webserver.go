package webserver

import (
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jonaz/ginlogrus"
	"github.com/jonaz/gograce"
	"github.com/olahol/melody"
	"github.com/rakyll/statik/fs"
	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/ca"
	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/handlers"
	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/helpers"
	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/models"
	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/models/persons"
	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/store"
	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/websocket"
)

type Webserver struct {
	Store            *store.Store
	Melody           *melody.Melody
	Config           *models.Config
	WebsocketHandler handlers.WebsocketHandler
	router           http.Handler
	CA               *ca.CA
}

func New(s *store.Store, conf *models.Config, wsh handlers.WebsocketHandler, m *melody.Melody, ca *ca.CA) *Webserver {
	return &Webserver{
		Store:            s,
		Config:           conf,
		WebsocketHandler: wsh,
		Melody:           m,
		CA:               ca,
	}
}

var sessionCookieKey = generateRandomKey(32)

func (ws *Webserver) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ws.router.ServeHTTP(w, req)
}

func (ws *Webserver) Init(requireAuth bool) *gin.Engine {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(gzip.Gzip(gzip.DefaultCompression))

	ws.initMelody(requireAuth)

	r.Use(ginlogrus.New(logrus.StandardLogger()))
	r.Use(cors.New(cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return true
		},
		MaxAge: 12 * time.Hour,
	}))

	if requireAuth {
		store := cookie.NewStore(sessionCookieKey)
		r.Use(sessions.Sessions("stampzilla-session", store))
		r.POST("/login", ws.handleLogin())
		r.POST("/register", ws.handleRegister())
		r.GET("/cert", ws.handleDownloadCert())
		r.GET("/logout", ws.handleLogout())
	}

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

	server.Handler = ws.Init(tlsConfig != nil)
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

func (ws *Webserver) initMelody(requireAuth bool) {
	// Setup melody
	ws.Melody.Upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws.Melody.HandleConnect(ws.handleConnect(requireAuth))
	ws.Melody.HandleMessage(ws.handleMessage())
	ws.Melody.HandleDisconnect(ws.handleDisconnect())
}

func cspMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Security-Policy", "worker-src 'self';")
		c.Next()
	}
}

func (ws *Webserver) handleConnect(requireAuth bool) func(s *melody.Session) {
	return func(s *melody.Session) {
		msg, err := models.NewMessage("server-info", models.ServerInfo{
			Name:       ws.Config.Name,
			UUID:       ws.Config.UUID,
			TLSPort:    ws.Config.TLSPort,
			Port:       ws.Config.Port,
			Init:       ws.Store.CountAdmins() < 1,
			AllowLogin: helpers.IsPrivateIP(s.Request.RemoteAddr),
		})
		if err != nil {
			logrus.Error(err)
			return
		}
		msg.WriteTo(s)

		// Require an identity, if we are on the secure socket
		secure, ok := s.Keys["secure"]
		if requireAuth && !ok {
			// Websockets are not allowed to relay any http status codes to the client script.
			// https://www.w3.org/TR/websockets/#feedback-from-the-protocol
			// So to signal the webgui that the user is unauthorized we have to use exit codes above 4000.
			// Error codes above 4000-49999 are reserved for private use.
			s.CloseWithMsg(melody.FormatCloseMessage(4001, "unauthorized"))
			logrus.Warn("4001 unauthorized")
			s.CloseWithMsg(melody.FormatCloseMessage(4001, "unauthorized"))
			return
		}

		proto, _ := s.Get(websocket.KeyProtocol.String())
		id, _ := s.Get(websocket.KeyID.String())

		// Add the connection to our list of connections
		ws.Store.AddOrUpdateConnection(id.(string), &models.Connection{
			Type:       proto.(string),
			RemoteAddr: s.Request.RemoteAddr,
			Attributes: s.Keys,
			Session:    s,
		})

		err = ws.WebsocketHandler.Connect(s, s.Request, s.Keys)
		if err != nil {
			logrus.Error(err)
			s.CloseWithMsg(melody.FormatCloseMessage(5000, "internal error"))
			return
		}

		// Mark the websocket as ready
		s.Set("ready", true)

		// Send ready message to the client
		readyInfo := models.ReadyInfo{}
		if secure != nil {
			readyInfo.Method = secure.(string)
		}

		if i, exists := s.Get("identity"); exists {
			// Try to add info about the logged in user
			readyInfo.User = ws.Store.GetPerson(i.(string))
		}

		msg, err = models.NewMessage("ready", readyInfo)
		if err != nil {
			logrus.Error(err)
			return
		}
		msg.WriteTo(s)
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

		ready, ok := s.Get("ready")
		if !ok || !ready.(bool) {
			logrus.Warnf("ignored incoming '%s' message (not ready) from %s", data.Type, data.FromUUID)
			return
		}

		go func() {
			defer func() {
				if r := recover(); r != nil {
					err, ok := r.(error)
					if !ok {
						err = fmt.Errorf("%s", r)
					}

					if len(data.Request) == 0 {
						logrus.Error(err)
						return
					}

					msg, err := models.NewMessage("failure", err.Error())
					if err != nil {
						logrus.Error(err)
					}
					msg.Request = data.Request
					err = msg.WriteTo(s)
					if err != nil {
						logrus.Error(err)
					}
				}
			}()
			resp, err := ws.WebsocketHandler.Message(s, data)

			// The message contains a request ID, so respond with the result
			if len(data.Request) > 0 {
				msg, e := models.NewMessage("success", resp)
				if e != nil {
					logrus.Error(e)
				}
				if err != nil {
					msg, e = models.NewMessage("failure", err.Error())
					if err != nil {
						logrus.Error(e)
					}
				}

				msg.Request = data.Request
				err := msg.WriteTo(s)
				if err != nil {
					logrus.Error(err)
				}
			}

			if err != nil {
				logrus.Error(err)
			}
		}()
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

func (ws *Webserver) handleDownloadCert() func(c *gin.Context) {
	return func(c *gin.Context) {
		// Only allowed with local addresses
		if !helpers.IsPrivateIP(c.Request.RemoteAddr) {
			c.AbortWithError(403, fmt.Errorf("only local access is allowed"))
			return
		}

		// Find the session (cookies)
		session := sessions.Default(c)
		if session.Get("id") == nil {
			c.AbortWithError(403, fmt.Errorf("no session"))
			return
		}

		// Find the user and check that we are allowed to login
		user := ws.Store.GetPerson(session.Get("id").(string))
		if user == nil || !user.AllowLogin {
			c.AbortWithError(403, fmt.Errorf("user not accepted"))
			return
		}

		// Get the client certificate
		cert, err := ws.CA.GetUserCertificate(user.UUID)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		header := c.Writer.Header()
		header["Content-Type"] = []string{"application/x-pkcs12"}
		header["Content-Disposition"] = []string{"attachment; filename= stampzilla-client-cert-" + user.Username + ".p12"}

		c.Writer.Write(cert)
	}
}

func (ws *Webserver) handleWs(m *melody.Melody) func(c *gin.Context) {
	return func(c *gin.Context) {
		keys := make(map[string]interface{})
		uuid := uuid.New().String()

		// Try to identify the client
		if c.Request.TLS != nil {
			certs := c.Request.TLS.PeerCertificates
			if len(certs) > 0 {
				keys["identity"] = certs[0].Subject.CommonName
				keys["secure"] = "cert"

				// Only accept X- headers from clients with a certificate
				if c.Request.Header.Get("X-UUID") != "" {
					uuid = c.Request.Header.Get("X-UUID")
				}
			} else if helpers.IsPrivateIP(c.Request.RemoteAddr) {
				// Check the cookie session, only allowed in with local access
				session := sessions.Default(c)
				if session.Get("username") != nil {
					keys["identity"] = session.Get("id")
					keys["username"] = session.Get("username")
					keys["secure"] = "session"
				}
			}
		}

		// Accept the requested protocol if known
		knownProtocols := map[string]bool{
			"node": true,
			"gui":  true,
		}
		proto := c.Request.Header.Get("Sec-WebSocket-Protocol")

		if !knownProtocols[proto] {
			logrus.Errorf("webserver: protocol \"%s\" not allowed", proto)
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		if proto != "" {
			c.Writer.Header().Set("Sec-WebSocket-Protocol", proto)
			keys[websocket.KeyProtocol.String()] = proto
		}

		if ws.Store.Connection(uuid) != nil {
			logrus.Errorf("Connection with same UUID already exists: %s", uuid)
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		keys[websocket.KeyID.String()] = uuid
		err := m.HandleRequestWithKeys(c.Writer, c.Request, keys)
		if err != nil {
			logrus.Errorf("webserver: %s", err.Error())
			return
		}
	}
}

func (ws *Webserver) handleLogin() func(c *gin.Context) {
	return func(c *gin.Context) {
		// Only allowed with local addresses
		if !helpers.IsPrivateIP(c.Request.RemoteAddr) {
			c.AbortWithStatus(403)
			return
		}

		user, err := ws.Store.ValidateLogin(c.PostForm("username"), c.PostForm("password"))
		if err != nil {
			c.String(http.StatusUnauthorized, err.Error())
			return
		}

		if !user.AllowLogin {
			c.String(http.StatusUnauthorized, "not allowed to login")
			return
		}

		err = ws.login(c, user)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
	}
}

func (ws *Webserver) login(c *gin.Context, user *persons.Person) error {
	session := sessions.Default(c)
	session.Set("username", user.Username)
	session.Set("is_admin", user.IsAdmin)
	session.Set("id", user.UUID)

	return session.Save()
}

func (ws *Webserver) Logout(uuid string) error {
	logrus.Error("Logout user ", uuid)
	connections := ws.Store.GetConnections()
	for _, c := range connections {
		for k, v := range c.Attributes {
			if k == "identity" && v == uuid {
				c.Session.CloseWithMsg(melody.FormatCloseMessage(4001, "unauthorized"))
			}
		}
	}
	return nil
}

func (ws *Webserver) handleRegister() func(c *gin.Context) {
	return func(c *gin.Context) {
		// Only allowed with local addresses
		if !helpers.IsPrivateIP(c.Request.RemoteAddr) {
			c.AbortWithStatus(403)
			return
		}

		// Only allow register if no admin exists
		if ws.Store.CountAdmins() > 0 {
			c.AbortWithStatus(403)
			return
		}

		if len(c.PostForm("username")) < 1 {
			c.String(http.StatusBadRequest, "username is to short, min 1 character")
			c.Abort()
			return
		}

		if len(c.PostForm("password")) < 8 {
			c.String(http.StatusBadRequest, "password is to short, min 8 characters")
			c.Abort()
			return
		}

		user := persons.PersonWithPasswords{
			NewPassword:    c.PostForm("password"),
			RepeatPassword: c.PostForm("password"),
			PersonWithPassword: persons.PersonWithPassword{
				Person: persons.Person{
					UUID:       uuid.New().String(),
					Name:       c.PostForm("username"),
					Username:   c.PostForm("username"),
					IsAdmin:    true,
					AllowLogin: true,
				},
			},
		}

		err := user.UpdatePassword()
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		err = ws.Store.AddOrUpdatePerson(user)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		err = ws.login(c, &user.Person)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
	}
}

func (ws *Webserver) handleLogout() func(c *gin.Context) {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		session.Clear()
		err := session.Save()
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
	}
}

func generateRandomKey(length int) []byte {
	k := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, k); err != nil {
		return nil
	}
	return k
}
