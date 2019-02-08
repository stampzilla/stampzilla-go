package store

import (
	"encoding/json"

	"github.com/sirupsen/logrus"
)

func (s *Store) AddOrUpdateServer(area, item string, state map[string]interface{}) {
	s.Lock()
	if s.Server[area] == nil {
		s.Server[area] = make(map[string]map[string]interface{})
	}
	if s.Server[area][item] == nil {
		s.Server[area][item] = make(map[string]interface{})
	}

	for k, v := range state {
		s.Server[area][item][k] = v
	}
	s.Unlock()

	s.runCallbacks("server")
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
