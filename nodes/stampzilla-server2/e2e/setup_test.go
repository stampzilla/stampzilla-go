package main

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/models"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/servermain"
	"github.com/stampzilla/stampzilla-go/pkg/node"
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

	// Server will tell the node its TLS port after successfull certificate request
	main.Config.TLSPort = secureURL[1]

	os.Setenv("STAMPZILLA_HOST", insecureURL[0])
	os.Setenv("STAMPZILLA_PORT", insecureURL[1])

	node := node.New("example")

	return main, node, func() {
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
	server.HTTPServer.Init()
	server.TLSServer.Init()

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
