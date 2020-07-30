package persons

import (
	"time"

	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models/devices"
)

type Person struct {
	UUID     string    `json:"uuid"`
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	LastSeen time.Time `json:"last_seen"`

	State devices.State `json:"state"`
}

type PersonWithPassword struct {
	Person
	Password string `json:"password"`
}

// Equal checks if 2 persons are equal
func (a *Person) Equal(b *PersonWithPassword) bool {
	if !a.State.Equal(b.State) { // this is most likely to not be equal so we check it first
		return false
	}
	if a.Name != b.Name {
		return false
	}
	if a.Email != b.Email {
		return false
	}
	if b.Password != "" {
		return false
	}
	if a.LastSeen != b.LastSeen {
		return false
	}

	return true
}
