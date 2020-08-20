package oauth

import (
	"net/http"
	"net/url"

	"github.com/davecgh/go-spew/spew"
	"github.com/gin-gonic/gin"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/go-session/session"
)

func AddRoutes(r *gin.Engine, srv *server.Server) {
	r.GET("/login", loginHandler)
	r.POST("/login", loginHandler)
	r.GET("/auth", authHandler)

	r.Any("/authorize", func(c *gin.Context) {
		// Startup a session
		store, err := session.Start(c.Request.Context(), c.Writer, c.Request)
		if err != nil {
			http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
			return
		}

		// Restore saved authorization parameters from session
		var form url.Values
		if v, ok := store.Get("ReturnUri"); ok {
			form = v.(url.Values)
		}
		c.Request.Form = form

		store.Delete("ReturnUri")
		store.Save()

		// Handover the request to the oauth provider lib that runs userAuthorizeHandler to check if the user is logged in
		err = srv.HandleAuthorizeRequest(c.Writer, c.Request)
		if err != nil {
			http.Error(c.Writer, err.Error(), http.StatusBadRequest)
		}
	})

	r.POST("/token", func(c *gin.Context) {
		spew.Dump(c.Request)
		srv.HandleTokenRequest(c.Writer, c.Request)
	})
}

func loginHandler(c *gin.Context) {
	store, err := session.Start(c.Request.Context(), c.Writer, c.Request)
	if err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
		return
	}

	if c.Request.Method == "POST" {
		if c.Request.Form == nil {
			if err := c.Request.ParseForm(); err != nil {
				http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		store.Set("instance", c.Request.Form.Get("instance"))
		store.Set("user_id", c.Request.Form.Get("username"))
		store.Save()

		c.Writer.Header().Set("Location", "/auth")
		c.Writer.WriteHeader(http.StatusFound)
		return
	}
	c.File("web/login.html")
}

func authHandler(c *gin.Context) {
	store, err := session.Start(nil, c.Writer, c.Request)
	if err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
		return
	}

	if _, ok := store.Get("user_id"); !ok {
		c.Writer.Header().Set("Location", "/login")
		c.Writer.WriteHeader(http.StatusFound)
		return
	}

	c.File("web/auth.html")
}
