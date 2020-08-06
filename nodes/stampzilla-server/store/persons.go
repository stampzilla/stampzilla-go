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

func (store *Store) AddOrUpdatePerson(p persons.PersonWithPasswords) error {
	store.Lock()
	store.Unlock()

	a := store.Persons.Get(p.UUID)
	if a != nil && a.Equal(p) {
		return nil
	}

	err := store.Persons.Add(p)
	if err != nil {
		return err
	}

	store.Persons.Save()
	store.runCallbacks("persons")

	return nil
}

func (store *Store) AddOrUpdatePersons(persons map[string]persons.PersonWithPasswords) error {
	store.RLock()
	previous := store.Persons.All()
	store.RUnlock()

	for id, person := range persons {
		person.UUID = id
		err := store.AddOrUpdatePerson(person)
		if err != nil {
			return err
		}

		delete(previous, id)
	}

	for uuid, _ := range previous {
		err := store.Persons.Delete(uuid)
		if err != nil {
			return err
		}
	}

	return nil
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
