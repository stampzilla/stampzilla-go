package store

type Request struct {
	Identity   string         `json:"identity"`
	Subject    RequestSubject `json:"subject"`
	Connection string         `json:"connection"`
	Type       string         `json:"type"`
	Version    string         `json:"version"`

	Approved chan bool `json:"-"`
}

type RequestSubject struct {
	CommonName         string        `json:"common_name,omitempty"`
	SerialNumber       string        `json:"serial_number,omitempty"`
	Country            []string      `json:"country,omitempty"`
	Organization       []string      `json:"organization,omitempty"`
	OrganizationalUnit []string      `json:"organizational_unit,omitempty"`
	Locality           []string      `json:"locality,omitempty"`
	Province           []string      `json:"province,omitempty"`
	StreetAddress      []string      `json:"street_address,omitempty"`
	PostalCode         []string      `json:"postal_code,omitempty"`
	Names              []interface{} `json:"names,omitempty"`
	// ExtraNames         []interface{} `json:"extra_names,omitempty"`
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
