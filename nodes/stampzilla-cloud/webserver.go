package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/acme/autocert"

	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-cloud/oauth"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models"
)

type Webserver struct {
	s *http.Server
}

func NewWebserver(pool *Pool) *Webserver {
	os.MkdirAll("./certs/acme", 0700)

	r := gin.Default()
	r.Use(func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": c.Errors[0].Error()})
		}
	})

	provider := oauth.New()
	oauth.AddRoutes(r, provider)

	r.Any("/webhook/:service", func(c *gin.Context) {
		a := mustValidateToken(provider, c)
		if a == nil {
			return
		}

		i, err := pool.GetByInstance(a.Instance)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		service := c.Param("service")
		i.ForwardRequest(service, c)
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
		s: s,
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
		}
	}

	logrus.Error("Got unexpected message")
}

func mustValidateToken(p *server.Server, c *gin.Context) *oauth.Authorization {
	t, err := p.ValidationBearerToken(c.Request)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return nil
	}

	var a oauth.Authorization
	if err := json.Unmarshal([]byte(t.GetUserID()), &a); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return nil
	}

	return &a
}
