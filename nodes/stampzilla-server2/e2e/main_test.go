package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/posener/wstest"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/models"
	"github.com/stretchr/testify/assert"
)

func TestDownloadCA(t *testing.T) {

	main, cleanup := setupServer(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "http://localhost/ca.crt", nil)
	w := httptest.NewRecorder()
	main.HTTPServer.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "application/x-x509-ca-cert", resp.Header.Get("Content-Type"))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, string(body), "BEGIN CERTIFICATE")
}

func TestInsecureWebsocket(t *testing.T) {

	main, cleanup := setupServer(t)
	defer cleanup()

	d := wstest.NewDialer(main.HTTPServer)
	d.Subprotocols = []string{"node"}
	c, _, err := d.Dial("ws://example.org/ws", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	msg := models.Message{}
	err = c.ReadJSON(&msg)
	if err != nil {
		t.Fatal(err)
	}

	serverInfo := &models.ServerInfo{}
	err = json.Unmarshal(msg.Body, &serverInfo)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "server-info", msg.Type)
	assert.Equal(t, "123", serverInfo.UUID)
	assert.Equal(t, "testserver", serverInfo.Name)
}
