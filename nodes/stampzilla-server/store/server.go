package store

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

func (store *Store) GetServerState() map[string]map[string]map[string]interface{} {
	store.RLock()
	defer store.RUnlock()
	return store.Server
}
