package store

import "github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models/persons"

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
