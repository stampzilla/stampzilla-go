package main

import (
	"fmt"
	"net/http"

	"github.com/RangelReale/osin"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func authorize(oauth2server *osin.Server) func(c *gin.Context) {
	return func(c *gin.Context) {
		resp := oauth2server.NewResponse()
		defer resp.Close()

		if ar := oauth2server.HandleAuthorizeRequest(resp, c.Request); ar != nil {

			// HANDLE LOGIN PAGE HERE
			if !handleLoginPage(ar, c.Writer, c.Request) {
				return
			}
			ar.UserData = struct{ Login string }{Login: "test"}
			ar.Authorized = true
			oauth2server.FinishAuthorizeRequest(resp, c.Request, ar)
		}
		if resp.IsError && resp.InternalError != nil {
			logrus.Errorf("ERROR: %#v\n", resp.InternalError)
		}
		osin.OutputJSON(resp, c.Writer, c.Request)
	}
}
func token(oauth2server *osin.Server) func(c *gin.Context) {
	return func(c *gin.Context) {
		resp := oauth2server.NewResponse()
		defer resp.Close()

		if ar := oauth2server.HandleAccessRequest(resp, c.Request); ar != nil {
			ar.Authorized = true
			oauth2server.FinishAccessRequest(resp, c.Request, ar)
		}
		if resp.IsError && resp.InternalError != nil {
			logrus.Errorf("ERROR: %#v\n", resp.InternalError)
		}
		osin.OutputJSON(resp, c.Writer, c.Request)
	}
}

func handleLoginPage(ar *osin.AuthorizeRequest, w http.ResponseWriter, r *http.Request) bool {
	r.ParseForm()
	if r.Method == "POST" && r.Form.Get("login") == "test" && r.Form.Get("password") == "test" {
		return true
	}

	w.Write([]byte("<html><body>"))

	w.Write([]byte(fmt.Sprintf("LOGIN %s (use test/test)<br/>", ar.Client.GetId())))
	w.Write([]byte(fmt.Sprintf("<form action=\"/authorize?%s\" method=\"POST\">", r.URL.RawQuery)))

	w.Write([]byte("Login: <input type=\"text\" name=\"login\" /><br/>"))
	w.Write([]byte("Password: <input type=\"password\" name=\"password\" /><br/>"))
	w.Write([]byte("<input type=\"submit\"/>"))

	w.Write([]byte("</form>"))

	w.Write([]byte("</body></html>"))

	return false
}
