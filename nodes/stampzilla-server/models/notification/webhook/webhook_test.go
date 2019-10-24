package webhook

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	sender := New(json.RawMessage("{\"method\": \"method1\"}"))

	assert.Equal(t, "method1", sender.Method)
}

func TestTrigger(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/webhook", req.URL.String())
		rw.Write([]byte(`OK`))
	}))
	defer server.Close()

	sender := New(json.RawMessage("{\"method\": \"PUT\"}"))
	err := sender.Trigger([]string{server.URL + "/webhook"}, "")

	assert.NoError(t, err)
}

func TestRelease(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/webhook", req.URL.String())
		rw.Write([]byte(`OK`))
	}))
	defer server.Close()

	sender := New(json.RawMessage("{\"method\": \"PUT\"}"))
	err := sender.Release([]string{server.URL + "/webhook"}, "")

	assert.NoError(t, err)
}

func TestDestinations(t *testing.T) {
	sender := New(json.RawMessage("{\"method\": \"PUT\"}"))
	d, err := sender.Destinations()
	assert.Nil(t, d)
	assert.Error(t, err)
}
