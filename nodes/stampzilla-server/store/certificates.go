package store

import "time"

type Certificate struct {
	Serial     string         `json:"serial"`
	Subject    RequestSubject `json:"subject"`
	CommonName string         `json:"commonName"`
	IsCA       bool           `json:"isCA"`
	Usage      []string       `json:"usage"`
	Revoked    bool           `json:"revoked"`
	Issued     time.Time      `json:"issued"`
	Expires    time.Time      `json:"expires"`

	Fingerprints map[string]string `json:"fingerprints"`
}

func (store *Store) GetCertificates() []Certificate {
	store.RLock()
	defer store.RUnlock()
	return store.Certificates
}

func (store *Store) UpdateCertificates(certs []Certificate) {
	store.Lock()
	store.Certificates = certs
	store.Unlock()

	store.runCallbacks("certificates")
}
