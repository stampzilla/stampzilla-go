package store

import (
	"encoding/json"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/models/devices"
)

func (s *Store) AddOrUpdateServer(area, item string, state devices.State) {
	s.Lock()
	if s.Server[area] == nil {
		s.Server[area] = make(map[string]devices.State)
	}
	if s.Server[area][item] == nil {
		s.Server[area][item] = make(devices.State)
	}

	if diff := s.Server[area][item].Diff(state); len(diff) != 0 {
		s.Server[area][item].MergeWith(diff)
		s.Unlock()
		s.runCallbacks("server")
		return
	}
	s.Unlock()
}

func (store *Store) GetServerStateAsJson() json.RawMessage {
	store.RLock()
	b, err := json.Marshal(store.Server)
	store.RUnlock()

	if err != nil {
		logrus.Errorf("Failed to marshal server state: %s", err.Error())
	}
	return b
}
