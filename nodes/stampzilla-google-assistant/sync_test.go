package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stampzilla/stampzilla-go/nodes/basenode"
	"github.com/stretchr/testify/assert"
)

func getData() string {
	return `
	{
		"nodeUuid.1": {
			"type": "lamp",
			"node": "nodeUuid",
			"id": "1",
			"name": "Test 1",
			"online": true,
			"tags": null
		},
		"nodeUuid.2": {
			"type": "lamp",
			"node": "nodeUuid",
			"id": "2",
			"name": "Test 2",
			"online": true,
			"tags": null
		}
	}
	`
}

func getDataFunc() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, getData())
	}
}

func fakeConfig(t *testing.T, h func(w http.ResponseWriter, r *http.Request)) (*basenode.Config, *nodeSpecificConfig, *httptest.Server) {

	ts := httptest.NewServer(http.HandlerFunc(h))

	t.Log(ts.URL)
	host, port, err := net.SplitHostPort(strings.TrimPrefix(ts.URL, "http://"))
	assert.NoError(t, err)

	config := &basenode.Config{
		Host: host,
	}
	ns := &nodeSpecificConfig{
		Port: port,
	}
	return config, ns, ts
}

func fetchURL(t *testing.T, method, u, body string) string {
	client := &http.Client{}
	var req *http.Request
	var err error
	if body != "" {
		req, err = http.NewRequest(method, u, strings.NewReader(body))
	} else {
		req, err = http.NewRequest(method, u, nil)
	}
	assert.NoError(t, err)
	res, err := client.Do(req)
	assert.NoError(t, err)
	bodyRes, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)
	res.Body.Close()
	return string(bodyRes)
}

func TestFetchDevices(t *testing.T) {
	config, ns, ts := fakeConfig(t, getDataFunc())
	defer ts.Close()
	devs, err := fetchDevices(config, ns)
	assert.NoError(t, err)
	dev1 := devs.ByID("nodeUuid.1")
	dev2 := devs.ByID("nodeUuid.2")
	if assert.NotNil(t, dev1) {
		assert.Equal(t, "1", dev1.Id)
		assert.Equal(t, "Test 1", dev1.Name)
	}
	if assert.NotNil(t, dev2) {
		assert.Equal(t, "2", dev2.Id)
		assert.Equal(t, "Test 2", dev2.Name)
	}
}
func TestSyncDevicesFromServer(t *testing.T) {
	config, ns, ts := fakeConfig(t, getDataFunc())
	defer ts.Close()

	success := syncDevicesFromServer(config, ns)
	t.Log("success:", success)
	assert.Equal(t, "nodeUuid.1", ns.Devices()[0].ID)
	assert.Equal(t, "nodeUuid.2", ns.Devices()[1].ID)
}
