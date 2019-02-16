package models

type Request struct {
	Version string `json:"version,omitempty"`
	Type    string `json:"type,omitempty"`
	CSR     string `json:"csr"`
}
