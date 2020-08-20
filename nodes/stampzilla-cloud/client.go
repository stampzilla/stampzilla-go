package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models"
)

type Client struct {
	Name string
	ID   string
	Conn *tls.Conn
	Pool *Pool

	requestID int
	requests  map[int]chan models.Message

	sync.RWMutex
}

func (cl *Client) ForwardRequest(service string, c *gin.Context) {
	defer c.Request.Body.Close()
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Request.ParseForm()

	var m *models.Message
	m, err = models.NewMessage("forwarded-request", models.ForwardedRequest{
		Method:     c.Request.Method,
		URL:        c.Request.URL,
		Header:     c.Request.Header,
		Body:       body,
		Form:       c.Request.Form,
		RemoteAddr: c.Request.RemoteAddr,
	})
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ch := make(chan models.Message)

	cl.Lock()
	m.Request = json.RawMessage(strconv.Itoa(cl.requestID))
	cl.requests[cl.requestID] = ch
	cl.requestID++
	cl.Unlock()

	m.WriteToWriter(cl.Conn)

	select {
	case resp := <-ch:
		switch resp.Type {
		case "failure":
			c.AbortWithError(http.StatusInternalServerError, fmt.Errorf(string(resp.Body)))
		case "success":
			c.Writer.Write(resp.Body)
		}
	case <-time.After(time.Second * 10):
		c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("timeout"))
	}
}
