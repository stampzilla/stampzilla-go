package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stampzilla/stampzilla-go/nodes/basenode"
	"github.com/stampzilla/stampzilla-go/pkg/hueemulator"
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

func fakeConfig(t *testing.T, h func(w http.ResponseWriter, r *http.Request)) (*basenode.Config, *nodeSpecific, *httptest.Server) {

	ts := httptest.NewServer(http.HandlerFunc(h))

	t.Log(ts.URL)
	host, port, err := net.SplitHostPort(strings.TrimPrefix(ts.URL, "http://"))
	assert.NoError(t, err)

	config := &basenode.Config{
		Host: host,
	}
	ns := &nodeSpecific{
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
	assert.Equal(t, "1", devs["nodeUuid.1"].Id)
	assert.Equal(t, "Test 1", devs["nodeUuid.1"].Name)
	assert.Equal(t, "2", devs["nodeUuid.2"].Id)
	assert.Equal(t, "Test 2", devs["nodeUuid.2"].Name)
}
func TestSyncDevicesFromServer(t *testing.T) {
	config, ns, ts := fakeConfig(t, getDataFunc())
	defer ts.Close()

	success := syncDevicesFromServer(config, ns)
	t.Log("success:", success)
	assert.Equal(t, 1, ns.Devices[0].Id)
	assert.Equal(t, 2, ns.Devices[1].Id)
}
func TestSetupHueHandlers(t *testing.T) {
	//config, ns, ts := fakeConfig(t)
	config, ns, ts := fakeConfig(t, func(w http.ResponseWriter, r *http.Request) {
		expextedUrls := []string{
			"/api/devices",
			"/api/nodes/nodeUuid/cmd/on/1",
			"/api/nodes/nodeUuid/cmd/off/1",
			"/api/nodes/nodeUuid/cmd/level/1/100.000000",
			"/api/nodes/nodeUuid/cmd/on/2",
			"/api/nodes/nodeUuid/cmd/off/2",
			"/api/nodes/nodeUuid/cmd/level/2/100.000000",
		}
		url := r.URL.String()

		fail := true
		for _, v := range expextedUrls {
			if url == v {
				fail = false
			}

		}
		if fail {
			t.Error("Unexpected URL called:", url)
		}
		fmt.Fprintln(w, getData())
	})
	defer ts.Close()

	hueemulator.SetLogger(os.Stdout)
	syncDevicesFromServer(config, ns)
	setupHueHandlers(ns)

	dummy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	dummy.Close()
	_, huePort, _ := net.SplitHostPort(strings.TrimPrefix(dummy.URL, "http://"))

	go func() {
		t.Error(hueemulator.ListenAndServe("127.0.0.1:" + huePort))
	}()

	time.Sleep(100 * time.Millisecond)

	body := fetchURL(t, "GET", "http://127.0.0.1:"+huePort+"/api/asdf/lights", "")

	//t.Log(string(body))
	// the device can be added in random order. Always the new one always get MaxId+1
	assert.Contains(t, body, `:{"state":{"on":false,"bri":0,"reachable":true},"type":"Dimmable light","name":"Test 1","modelid":"LWB014","manufacturername":"Philips","uniqueid"`)
	assert.Contains(t, body, `:{"state":{"on":false,"bri":0,"reachable":true},"type":"Dimmable light","name":"Test 2","modelid":"LWB014","manufacturername":"Philips","uniqueid"`)
	//assert.Equal(t, data, strings.TrimSpace(string(body)))

	on := fetchURL(t, "PUT", "http://127.0.0.1:"+huePort+"/api/asdf/lights/1/state", `{"on":true}`)
	t.Log(on)
	assert.Equal(t, `[{"success":{"/lights/1/state/on":true}}]`, on)
	off := fetchURL(t, "PUT", "http://127.0.0.1:"+huePort+"/api/asdf/lights/1/state", `{"on":false}`)
	t.Log(off)
	assert.Equal(t, `[{"success":{"/lights/1/state/on":false}}]`, off)

	level := fetchURL(t, "PUT", "http://127.0.0.1:"+huePort+"/api/asdf/lights/1/state", `{"bri":255}`)
	t.Log(level)
}
