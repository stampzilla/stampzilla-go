package store

import (
	"sync/atomic"
	"testing"

	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models/devices"
	"github.com/stretchr/testify/assert"
)

func TestAddOrUpdateServer(t *testing.T) {
	store := &Store{
		Server: make(map[string]map[string]devices.State),
	}

	i := int64(0)
	store.OnUpdate(func(str string, s *Store) error {
		atomic.AddInt64(&i, 1)
		assert.Equal(t, "server", str)
		return nil
	})

	store.AddOrUpdateServer("area", "item", devices.State{
		"state1": true,
	})
	store.AddOrUpdateServer("area", "item", devices.State{
		"state1": true,
	})
	store.AddOrUpdateServer("area", "item", devices.State{
		"state2": true,
	})

	assert.Equal(t, int64(2), i)
}

func TestGetServerStateAsJson(t *testing.T) {
	store := &Store{
		Server: make(map[string]map[string]devices.State),
	}
	store.AddOrUpdateServer("area", "item", devices.State{
		"state1": true,
	})
	data := store.GetServerStateAsJson()
	assert.Contains(t, string(data), "state1")
}
