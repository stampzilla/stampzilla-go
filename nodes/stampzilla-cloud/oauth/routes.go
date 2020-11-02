package oauth

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/sirupsen/logrus"
)

func AddRoutes(r *gin.Engine, srv *server.Server, verifyUser func(instance, username, password string) (*Authorization, error)) {
	r.Any("/authorize", authorizationHandler(srv, verifyUser))

	r.OPTIONS("/token", func(c *gin.Context) {})
	r.POST("/token", func(c *gin.Context) {
		srv.HandleTokenRequest(c.Writer, c.Request)
	})
}

func authorizationHandler(srv *server.Server, verifyUser func(instance, username, password string) (*Authorization, error)) func(c *gin.Context) {
	return func(c *gin.Context) {
		if c.Request.Method == "POST" {
			if c.Request.Form == nil {
				if err := c.Request.ParseForm(); err != nil {
					c.HTML(http.StatusBadRequest, "login.html", gin.H{
						"errorClass": "is-invalid",
						"error":      err.Error(),
					})
					return
				}
			}

			authorization, err := verifyUser(c.Request.Form.Get("instance"), c.Request.Form.Get("username"), c.Request.Form.Get("password"))
			if err != nil {
				c.HTML(http.StatusBadRequest, "login.html", gin.H{
					"errorClass": "is-invalid",
					"error":      err.Error(),
					"instance":   c.Request.Form.Get("instance"),
					"username":   c.Request.Form.Get("username"),
				})
				return
			}
			encoded, err := json.Marshal(authorization)
			if err != nil {
				c.HTML(http.StatusBadRequest, "login.html", gin.H{
					"errorClass": "is-invalid",
					"error":      err.Error(),
					"instance":   c.Request.Form.Get("instance"),
					"username":   c.Request.Form.Get("username"),
				})
				return
			}

			// Add the information to the request context
			req := c.Request.WithContext(context.WithValue(c.Request.Context(), "authorization", encoded))

			// Handover the request to the oauth provider lib that runs userAuthorizeHandler to check if the user is logged in
			err = srv.HandleAuthorizeRequest(c.Writer, req)
			if err != nil {
				logrus.Warn("Authorization failed:", err)
				c.HTML(http.StatusBadRequest, "login.html", gin.H{
					"errorClass": "is-invalid",
					"error":      err.Error(),
					"instance":   c.Request.Form.Get("instance"),
					"username":   c.Request.Form.Get("username"),
				})
				return
			}
		} else {
			c.HTML(200, "login.html", gin.H{})
		}
	}
}
