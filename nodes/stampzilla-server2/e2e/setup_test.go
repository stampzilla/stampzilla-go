package main

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/models"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/servermain"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/store"
)

func makeRequest(t *testing.T, handler http.Handler, method, url string, body io.Reader) *httptest.ResponseRecorder {
	req := httptest.NewRequest("GET", "http://localhost/ca.crt", body)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	return w
}

func setupServer(t *testing.T) (*servermain.Main, func()) {
	config := &models.Config{
		UUID: "123",
		Name: "testserver",
	}
	store := store.New()
	server := servermain.New(config, store)

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
