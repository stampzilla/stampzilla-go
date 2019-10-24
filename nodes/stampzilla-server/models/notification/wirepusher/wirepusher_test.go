package wirepusher

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	sender := New(json.RawMessage("{\"title\": \"title1\", \"type\": \"type1\", \"action\": \"action1\"}"))

	assert.Equal(t, "title1", sender.Title)
	assert.Equal(t, "type1", sender.Type)
	assert.Equal(t, "action1", sender.Action)
}

func TestTrigger(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/send?action=action1&id=deviceId1&message=test+body&title=title1+-+Triggered&type=type1", req.URL.String())
		rw.Write([]byte(`OK`))
	}))
	defer server.Close()

	sender := New(json.RawMessage("{\"title\": \"title1\", \"type\": \"type1\", \"action\": \"action1\"}"))
	sender.url = server.URL + "/send"
	err := sender.Trigger([]string{"deviceId1"}, "test body")

	assert.NoError(t, err)
}

func TestRelease(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/send?action=action1&id=deviceId1&message=test+body&title=title1+-+Released&type=type1", req.URL.String())
		rw.Write([]byte(`OK`))
	}))
	defer server.Close()

	sender := New(json.RawMessage("{\"title\": \"title1\", \"type\": \"type1\", \"action\": \"action1\"}"))
	sender.url = server.URL + "/send"
	err := sender.Release([]string{"deviceId1"}, "test body")

	assert.NoError(t, err)
}

func TestDestinations(t *testing.T) {
	sender := New(json.RawMessage("{\"title\": \"title1\", \"type\": \"type1\", \"action\": \"action1\"}"))
	d, err := sender.Destinations()
	assert.Nil(t, d)
	assert.Error(t, err)
}
