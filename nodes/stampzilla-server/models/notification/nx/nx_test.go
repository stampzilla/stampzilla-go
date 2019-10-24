package nx

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	sender := New(json.RawMessage("{\"server\": \"server1\", \"username\": \"user1\", \"password\": \"pass1\"}"))

	assert.Equal(t, "server1", sender.Server)
	assert.Equal(t, "user1", sender.Username)
	assert.Equal(t, "pass1", sender.Password)
}
