package email

import (
	"encoding/json"
	"net/smtp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	sender := New(json.RawMessage("{\"server\": \"server1\", \"port\": 123, \"from\": \"from1\", \"password\": \"pass1\"}"))

	assert.Equal(t, "server1", sender.Server)
	assert.Equal(t, 123, sender.Port)
	assert.Equal(t, "from1", sender.From)
	assert.Equal(t, "pass1", sender.Password)
}

func TestTrigger(t *testing.T) {
	f, r := mockSend(nil)
	sender := &EmailSender{send: f}
	err := sender.Trigger([]string{"me@example.com"}, "Hello World")

	assert.NoError(t, err)
	assert.Equal(t, "From: \nSubject: stampzilla - Triggered\n\nHello World", string(r.msg))
}

func TestRelease(t *testing.T) {
	f, r := mockSend(nil)
	sender := &EmailSender{send: f}
	err := sender.Release([]string{"me@example.com"}, "Hello World")

	assert.NoError(t, err)
	assert.Equal(t, "From: \nSubject: stampzilla - Released\n\nHello World", string(r.msg))
}

func TestDestinations(t *testing.T) {
	sender := &EmailSender{}
	d, err := sender.Destinations()
	assert.Nil(t, d)
	assert.Error(t, err)
}

func mockSend(errToReturn error) (func(string, smtp.Auth, string, []string, []byte) error, *emailRecorder) {
	r := new(emailRecorder)
	return func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		*r = emailRecorder{addr, a, from, to, msg}
		return errToReturn
	}, r
}

type emailRecorder struct {
	addr string
	auth smtp.Auth
	from string
	to   []string
	msg  []byte
}
