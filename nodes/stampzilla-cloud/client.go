package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models"
)

type Client struct {
	Name     string
	Instance string
	ID       string
	Phrase   string
	Conn     *tls.Conn
	Pool     *Pool

	requestID int
	requests  map[int]chan models.Message

	sync.RWMutex
}

func (cl *Client) Disconnect() {
	cl.Conn.Close()
}
func (cl *Client) Authorize(username, password string) (userID string, err error) {
	var m *models.Message
	m, err = models.NewMessage("authorize-request", models.AuthorizeRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		return "", err
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
			return "", fmt.Errorf(string(resp.Body))
		case "success":
			var clientID string
			err := json.Unmarshal(resp.Body, &clientID)
			return clientID, err
		}
	case <-time.After(time.Second * 10):
		return "", fmt.Errorf("timeout")
	}

	return "", fmt.Errorf("should never happen")
}

func (cl *Client) ForwardRequest(service string, c *gin.Context) {
	dump, err := httputil.DumpRequest(c.Request, true)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	var m *models.Message
	m, err = models.NewMessage("forwarded-request", models.ForwardedRequest{
		Dump:       dump,
		Service:    service,
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
			var body string
			err := json.Unmarshal(resp.Body, &body)
			if err != nil {
				logrus.Error(err)
				return
			}

			raw, err := base64.StdEncoding.DecodeString(body)
			if err != nil {
				logrus.Error(err)
				return
			}

			b := bytes.NewBuffer(raw)
			rd := bufio.NewReader(b)
			resp, err := http.ReadResponse(rd, c.Request)
			if err != nil {
				logrus.Error(err)
				return
			}

			for k, v := range resp.Header {
				for _, v2 := range v {
					c.Writer.Header().Add(k, v2)
				}
			}
			c.Writer.WriteHeader(resp.StatusCode)
			io.Copy(c.Writer, resp.Body)

			//c.Writer.Header()["Content-Type"] = []string{"application/json"}
			//c.Writer.Write(resp.Body)
		}
	case <-time.After(time.Second * 10):
		c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("timeout"))
	}
}
