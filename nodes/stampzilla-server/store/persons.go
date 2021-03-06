package store

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/models/persons"
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

func (store *Store) CountAdmins() int {
	store.RLock()
	defer store.RUnlock()

	persons := store.Persons.All()
	cnt := 0

	for _, p := range persons {
		if p.IsAdmin {
			cnt = cnt + 1
		}
	}

	return cnt
}

func (store *Store) AddOrUpdatePerson(p persons.PersonWithPasswords) error {
	store.Lock()
	store.Unlock()

	a := store.Persons.Get(p.UUID)
	if a != nil && a.Equal(p) {
		return nil
	}

	// Demotion of an admin, check that it is at least one admin left
	if a != nil && a.IsAdmin && !p.IsAdmin && store.CountAdmins() == 1 {
		return fmt.Errorf("not allowed to remove the last admin")
	}

	err := store.Persons.Add(p)
	if err != nil {
		return err
	}

	store.Persons.Save()
	store.runCallbacks("persons")

	// Logout the demoted user from our server
	if a != nil && a.IsAdmin && !p.IsAdmin {
		for _, callback := range store.onUserDemote {
			if err := callback(p.UUID); err != nil {
				logrus.Error("store user demote callback: ", err)
			}
		}
	}

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

	if len(previous) > 0 {
		for uuid, person := range previous {
			if person.IsAdmin && store.CountAdmins() == 1 {
				return fmt.Errorf("not allowed to remove the last admin")
			}

			err := store.Persons.Delete(uuid)
			if err != nil {
				return err
			}

			// Logout the deleted user from our server
			for _, callback := range store.onUserDemote {
				if err := callback(uuid); err != nil {
					logrus.Error("store user demote callback: ", err)
				}
			}
		}

		store.Persons.Save()
		store.runCallbacks("persons")
	}

	return nil
}

func (store *Store) ValidateLogin(username, password string) (*persons.Person, error) {
	store.RLock()
	defer store.RUnlock()

	if len(username) < 1 {
		return nil, fmt.Errorf("no username provided")
	}

	p := store.Persons.GetByUsernameWithPassowrd(username)
	if p == nil {
		return nil, fmt.Errorf("wrong username or password")
	}

	err := p.CheckPassword(password)
	if err != nil {
		return nil, fmt.Errorf("wrong username or password")
	}

	return &p.Person, nil
}
