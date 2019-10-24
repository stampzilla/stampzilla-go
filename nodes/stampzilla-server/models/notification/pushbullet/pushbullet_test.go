package pushbullet

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	sender := New(json.RawMessage("{\"token\": \"token1\"}"))

	assert.Equal(t, "token1", sender.Token)
}
