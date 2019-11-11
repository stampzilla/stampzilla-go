package main

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/servermain"
	"github.com/stampzilla/stampzilla-go/pkg/node"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/publicsuffix"
)

func makeRequest(t *testing.T, handler http.Handler, method, url string, body io.Reader) *httptest.ResponseRecorder {
	req := httptest.NewRequest("GET", "http://localhost/ca.crt", body)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	return w
}

func setupWebsocketTest(t *testing.T) (*servermain.Main, *node.Node, func()) {
	main, cleanup := setupServer(t)
	insecure := httptest.NewServer(main.HTTPServer)

	secure := httptest.NewUnstartedServer(main.TLSServer)
	secure.TLS = main.TLSConfig()
	secure.StartTLS()

	insecureURL := strings.Split(strings.TrimPrefix(insecure.URL, "http://"), ":")
	secureURL := strings.Split(strings.TrimPrefix(secure.URL, "https://"), ":")

	// Server will tell the node its TLS port after successful certificate request
	main.Config.TLSPort = secureURL[1]

	os.Setenv("STAMPZILLA_HOST", insecureURL[0])
	os.Setenv("STAMPZILLA_PORT", insecureURL[1])

	node := node.New("example")

	ctx, cancel := context.WithCancel(context.Background())
	main.Store.Logic.Start(ctx)
	main.Store.Scheduler.Start(ctx)

	return main, node, func() {
		cancel()
		cleanup()
		insecure.Close()
		secure.Close()
	}
}

func setupServer(t *testing.T) (*servermain.Main, func()) {
	config := &models.Config{
		UUID: "123",
		Name: "testserver",
	}
	server := servermain.New(config)

	prevDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir, err := ioutil.TempDir("", "e2etest")
	if err != nil {
		log.Fatal(err)
	}
	os.Chdir(dir)

	server.Init()
	server.HTTPServer.Init(false)
	server.TLSServer.Init(true)

	cleanUp := func() {
		os.Chdir(prevDir)
		err := os.RemoveAll(dir) // clean up
		if err != nil {
			t.Fatal(err)
		}
	}
	return server, cleanUp
}

func waitFor(t *testing.T, timeout time.Duration, msg string, ok func() bool) {
	end := time.Now().Add(timeout)
	for {
		if end.Before(time.Now()) {
			t.Errorf("timeout waiting for: %s", msg)
			return
		}
		time.Sleep(10 * time.Millisecond)
		if ok() {
			return
		}
	}
}

func acceptCertificateRequest(t *testing.T, main *servermain.Main) {
	go func() {
		waitFor(t, 2*time.Second, "nodes should be 1", func() bool {
			return len(main.Store.GetRequests()) == 1
		})
		r := main.Store.GetRequests()
		main.Store.AcceptRequest(r[0].Connection)
	}()
}

func login(t *testing.T, main *servermain.Main, d *websocket.Dialer) {
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		log.Fatal(err)
	}

	var body bytes.Buffer
	mp := multipart.NewWriter(&body)
	username, err := mp.CreateFormField("username")
	assert.NoError(t, err)
	_, err = io.Copy(username, strings.NewReader("test-user"))
	assert.NoError(t, err)
	password, err := mp.CreateFormField("password")
	assert.NoError(t, err)
	_, err = io.Copy(password, strings.NewReader("test-pass"))
	assert.NoError(t, err)
	mp.Close()

	req := httptest.NewRequest("POST", "http://example.org/login", &body)
	req.Header.Set("Content-Type", mp.FormDataContentType())

	w := httptest.NewRecorder()

	main.TLSServer.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	resp := w.Result()
	url, err := url.Parse("http://example.org/login")
	assert.NoError(t, err)

	jar.SetCookies(url, resp.Cookies())
	d.Jar = jar
}
