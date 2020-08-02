package store

import (
	"fmt"

	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models/persons"
	"golang.org/x/crypto/bcrypt"
)

func (store *Store) GetPersons() persons.PersonMap {
	store.RLock()
	defer store.RUnlock()
	return store.Persons.All()
}

func (store *Store) GetPerson(uuid string) *persons.Person {
	store.RLock()
	defer store.RUnlock()
	return store.Persons.Get(uuid)
}

func (store *Store) AddOrUpdatePerson(p *persons.PersonWithPassword) {
	store.Lock()
	store.Unlock()

	a := store.Persons.Get(p.UUID)
	if a != nil && a.Equal(p) {
		return
	}

	store.Persons.Add(p)

	store.runCallbacks("persons")
}

func (store *Store) ValidateLogin(email, password string) (*persons.Person, error) {
	store.RLock()
	defer store.RUnlock()

	p := store.Persons.GetByEmailWithPassowrd(email)
	if p == nil {
		return nil, fmt.Errorf("wrong username or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(p.Password), []byte(password)); err != nil {
		return nil, fmt.Errorf("wrong username or password")
	}

	return &p.Person, nil
}
