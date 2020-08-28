package main

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"net/http"
	"os"
	"strconv"

	"github.com/foolin/goview/supports/ginview"
	"github.com/gin-gonic/gin"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/sirupsen/logrus"
	"github.com/tyler-smith/go-bip39/wordlists"
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
	r.HTMLRender = ginview.Default()
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
		a := mustValidateToken(provider, c)
		if a == nil {
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
			return
		}
	}

	logrus.Error("Got unexpected message")

	//spew.Dump(resp)
	//spew.Dump(c.requests)
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
