package models

import (
	"testing"

	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/models/devices"
	"github.com/stretchr/testify/assert"
)

func TestSetAlias(t *testing.T) {
	n := &Node{}
	id, _ := devices.NewIDFromString("1.1")
	n.SetAlias(id, "alias")
	assert.Equal(t, "alias", n.Alias(id))
}
