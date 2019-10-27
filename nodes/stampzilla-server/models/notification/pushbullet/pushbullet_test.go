package pushbullet

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	sender := New(json.RawMessage("{\"token\": \"token1\"}"))

	assert.Equal(t, "token1", sender.Token)
}

func TestTrigger(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/v2/pushes", req.URL.String())
		req.ParseForm()

		b, err := ioutil.ReadAll(req.Body)
		assert.NoError(t, err)

		assert.Equal(t, "{\"body\":\"booody\",\"device_iden\":\"pushbullet-id\",\"title\":\"stampzilla\",\"type\":\"note\"}", string(b))

		rw.Write([]byte(`OK`))
	}))
	defer server.Close()

	sender := New(json.RawMessage("{\"token\": \"token1\"}"))
	sender.server = server.URL
	err := sender.Trigger([]string{"pushbullet-id"}, "booody")

	assert.NoError(t, err)
}

func TestRelease(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Fail(t, "Should not call the api server")
	}))
	defer server.Close()

	sender := New(json.RawMessage("{\"token\": \"token1\"}"))
	sender.server = server.URL
	err := sender.Release([]string{"pushbullet-id"}, "booody")

	assert.Error(t, err)
}

func TestDestinations(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/v2/devices", req.URL.String())

		rw.Write([]byte(`{
  "accounts": [],
  "blocks": [],
  "channels": [],
  "chats": [],
  "clients": [],
  "contacts": [],
  "devices": [
    {
      "active": true,
      "iden": "id1",
      "created": 1516912020.3395772,
      "modified": 1571173323.7160761,
      "type": "android",
      "kind": "android",
      "nickname": "Sony Xperia",
      "manufacturer": "Sony",
      "model": "F5321",
      "app_version": 256,
      "fingerprint": "{\"android_id\":\"android_id_1\"}",
      "push_token": "pushtoken1",
      "pushable": true,
      "has_sms": true,
      "has_mms": true,
      "icon": "phone",
      "key_fingerprint": "blabla1",
      "remote_files": "disabled"
    },
    {
      "active": true,
      "iden": "id2",
      "created": 1571061309.753149,
      "modified": 1571126296.00051,
      "type": "android",
      "kind": "android",
      "nickname": "Galaxy S10e",
      "generated_nickname": true,
      "manufacturer": "samsung",
      "model": "SM-G970F",
      "app_version": 256,
      "fingerprint": "{\"android_id\":\"android_id_2\"}",
      "push_token": "pushtoken2",
      "pushable": true,
      "has_sms": true,
      "has_mms": true,
      "icon": "phone",
      "key_fingerprint": "blabla2",
      "remote_files": "disabled"
    },
    {
      "active": true,
      "iden": "id3",
      "created": 1516911936.359554,
      "modified": 1571125718.47128,
      "type": "chrome",
      "kind": "chrome",
      "nickname": "Google Chrome",
      "manufacturer": "Google",
      "model": "Chrome",
      "app_version": 339,
      "pushable": true,
      "icon": "browser",
      "key_fingerprint": "blabla"
    }
  ],
  "grants": [],
  "pushes": [],
  "profiles": [],
  "subscriptions": [],
  "texts": []
}`))
	}))
	defer server.Close()

	sender := New(json.RawMessage("{\"token\": \"token1\"}"))
	sender.server = server.URL
	d, err := sender.Destinations()
	assert.Equal(t, map[string]string{"id3": "Google Chrome", "id1": "Sony Xperia", "id2": "Galaxy S10e"}, d)
	assert.NoError(t, err)
}
