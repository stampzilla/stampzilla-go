package main

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/foolin/goview/supports/ginview"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/tyler-smith/go-bip39/wordlists"
	"golang.org/x/crypto/acme/autocert"

	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-cloud/oauth"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-cloud/websockets"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models"
)

type Webserver struct {
	s   *http.Server
	hub *websockets.Hub
}

func NewWebserver(pool *Pool) *Webserver {
	os.MkdirAll("./certs/acme", 0700)

	hub := websockets.NewHub()

	r := gin.Default()
	r.HTMLRender = ginview.Default()

	config := cors.DefaultConfig()
	config.AllowHeaders = []string{"Authorization"}
	config.AllowAllOrigins = true
	r.Use(cors.New(config))

	r.StaticFile("/tos", "./views/tos.html")
	r.StaticFile("/privacy", "./views/privacy.html")

	r.Use(func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": c.Errors[0].Error()})
		}
	})

	verifyUser := func(instance, username, password string) (*oauth.Authorization, error) {
		i, err := pool.GetByInstance(instance)
		if err != nil {
			return nil, fmt.Errorf("invalid credentials")
		}

		userID, err := i.Authorize(username, password)

		if err != nil {
			logrus.WithFields(logrus.Fields{
				"instance": instance,
				"username": username,
				"error":    err,
			}).Info("Failed to authorize user")
			return nil, fmt.Errorf("invalid credentials")
		}

		return &oauth.Authorization{
			ClientID: i.ID,
			UserID:   userID,
		}, nil
	}

	provider := oauth.New()
	oauth.AddRoutes(r, provider, verifyUser)

	r.GET("robots.txt", func(c *gin.Context) {
		c.File("views/robots.txt")
	})

	r.GET("/identify/:instance", func(c *gin.Context) {
		i, err := pool.GetByInstance(c.Param("instance"))
		if err != nil {
			// If the instance is missing, just answer with a word from a wordlist
			wordlist := wordlists.English

			h := fnv.New32a()
			h.Write([]byte(c.Param("instance")))
			word := int(h.Sum32()) % len(wordlist)

			c.JSON(200, gin.H{
				"name":   wordlist[word],
				"phrase": "",
			})
			return
		}

		c.JSON(200, gin.H{
			"name":   i.Name,
			"phrase": i.Phrase,
		})
	})

	r.Any("/webhook/:service", func(c *gin.Context) {
		a, status, err := mustValidateToken(provider, c)
		if a == nil {
			c.AbortWithError(status, err)
			return
		}

		i, err := pool.GetByID(a.ClientID)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		service := c.Param("service")
		i.ForwardRequest(service, c)
	})

	r.GET("/app/ws", func(c *gin.Context) {
		conn, err := wsupgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			logrus.Warnf("Failed to set websocket upgrade: %+v", err)
			return
		}

		a, _, err := mustValidateToken(provider, c)
		if err != nil {
			logrus.Error(err)
			conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(4001, ""))
			conn.Close()
			return
		}

		conn.WriteMessage(websocket.TextMessage, []byte("{\"type\":\"ready\"}"))

		callback := func(msg []byte) error {
			socketClient, err := pool.GetByID(a.ClientID)
			if err != nil {
				logrus.Error(err)
				return err
			}

			_, err = socketClient.Conn.Write(msg)
			if err != nil {
				logrus.Error(err)
				return err
			}
			return nil
		}

		wsClient := websockets.ServeWs(hub, conn, a, callback)

		socketClient, err := pool.GetByID(a.ClientID)
		if err == nil {
			wsClient.Send(socketClient.nodes)
			wsClient.Send(socketClient.devices)
		}
	})

	// ACME (Lets encrypt)
	m := &autocert.Manager{
		Cache:      autocert.DirCache("./certs/acme"),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist("alvalyckan.stamp.se"),
	}
	s := &http.Server{
		Addr:      ":https",
		TLSConfig: m.TLSConfig(),
		Handler:   r,
	}

	return &Webserver{
		s:   s,
		hub: hub,
	}
}

func (w *Webserver) Start() {
	logrus.Info("Start public TLS webserver")
	err := w.s.ListenAndServeTLS("", "")
	logrus.Fatal(err)
}

func (w *Webserver) HandleResponse(resp models.Message, c *Client) {
	if n, err := strconv.Atoi(string(resp.Request)); err == nil {
		c.RLock()
		ch, ok := c.requests[n]
		c.RUnlock()

		if ok {
			go func() {
				ch <- resp
			}()
			return
		}
	}

	logrus.Error("Got unexpected message: ", resp.Type)

	//spew.Dump(resp)
	//spew.Dump(c.requests)
}

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func getToken(r *http.Request) (token string, ok bool) {
	auth := r.Header.Get("Authorization")
	prefix := "Bearer "

	if auth != "" && strings.HasPrefix(auth, prefix) {
		token = auth[len(prefix):]
	} else {
		token = r.FormValue("access_token")
	}

	ok = token != ""
	return
}

func mustValidateToken(p *server.Server, c *gin.Context) (*oauth.Authorization, int, error) {
	accessToken, ok := getToken(c.Request)
	if !ok {
		return nil, http.StatusBadRequest, fmt.Errorf("no token provided")
	}

	ti, err := p.Manager.LoadAccessToken(c.Request.Context(), accessToken)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	var a oauth.Authorization
	if err := json.Unmarshal([]byte(ti.GetUserID()), &a); err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return &a, 0, nil
}
