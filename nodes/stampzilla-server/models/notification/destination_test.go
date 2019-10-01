package notification

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEqual(t *testing.T) {
	d1 := &Destination{
		Name: "name",
	}
	d2 := &Destination{
		Name: "name",
	}
	assert.True(t, d1.Equal(d2))
}

func TestAddDestination(t *testing.T) {
	dests := NewDestinations()
	d1 := &Destination{
		Name: "name",
		UUID: "uuid",
	}
	dests.Add(d1)
	assert.Equal(t, "name", dests.Get("uuid").Name)
}
func TestRemoveDestination(t *testing.T) {
	dests := NewDestinations()
	d1 := &Destination{
		Name: "name",
		UUID: "uuid",
	}
	dests.Add(d1)
	dests.Remove("uuid")
	assert.Len(t, dests.All(), 0)
}
