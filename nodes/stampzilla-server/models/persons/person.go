package persons

import (
	"fmt"
	"time"

	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models/devices"
	"golang.org/x/crypto/bcrypt"
)

type Person struct {
	UUID       string `json:"uuid"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	AllowLogin bool   `json:"allow_login"`
	IsAdmin    bool   `json:"is_admin"`

	LastSeen time.Time `json:"last_seen"`

	State devices.State `json:"state"`
}

type PersonWithPassword struct {
	Person
	Password string `json:"password"`
}

type PersonWithPasswords struct {
	PersonWithPassword
	NewPassword    string `json:"new_password,omitempty"`
	RepeatPassword string `json:"repeat_password,omitempty"`
}

// Equal checks if 2 persons are equal
func (a Person) Equal(b PersonWithPasswords) bool {
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
	if a.AllowLogin != b.AllowLogin {
		return false
	}
	if a.IsAdmin != b.IsAdmin {
		return false
	}
	if b.NewPassword != "" {
		return false
	}

	return true
}

func (a *PersonWithPasswords) UpdatePassword() error {
	if a.NewPassword != "" && a.NewPassword != a.RepeatPassword {
		return fmt.Errorf("repeat password does not match the new password")
	}

	// Change password
	if a.NewPassword != "" && a.NewPassword == a.RepeatPassword {
		hash, err := bcrypt.GenerateFromPassword([]byte(a.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to generate password hash: %s", err)
		}

		a.Password = string(hash)
	}

	a.NewPassword = ""
	a.RepeatPassword = ""

	return nil
}
