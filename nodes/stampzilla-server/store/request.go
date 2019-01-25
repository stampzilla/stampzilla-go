package store

type Request struct {
	Identity   string `json:"identity"`
	Connection string `json:"connection"`
	Type       string `json:"type"`
	Version    string `json:"version"`

	Approved chan bool `json:"-"`
}

func (store *Store) GetRequests() []Request {
	store.RLock()
	defer store.RUnlock()
	return store.Requests
}

func (store *Store) AddRequest(r Request) {
	store.Lock()
	store.Requests = append(store.Requests, r)
	store.Unlock()

	store.runCallbacks("requests")
}

func (store *Store) RemoveRequest(c string, approved bool) {
	store.Lock()
	for i, r := range store.Requests {
		if r.Connection == c {
			if r.Approved != nil {
				if approved {
					r.Approved <- true
				}
				close(r.Approved)
				r.Approved = nil
			}
			store.Requests = append(store.Requests[:i], store.Requests[i+1:]...)
		}
	}
	store.Unlock()

	store.runCallbacks("requests")
}

func (store *Store) AcceptRequest(c string) {
	store.RemoveRequest(c, true)
}
