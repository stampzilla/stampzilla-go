package persons

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type PersonMapWithPassword map[string]*PersonWithPassword
type PersonMap map[string]*Person

type List struct {
	persons PersonMapWithPassword
	sync.RWMutex
}

func NewList() List {
	return List{
		persons: make(map[string]*PersonWithPassword),
	}
}

// Add adds a person to the list
func (l *List) Add(p *PersonWithPassword) {
	l.Lock()
	l.persons[p.UUID] = p
	l.Unlock()
}

// All get all persons
func (l *List) All() PersonMap {
	l.RLock()
	defer l.RUnlock()

	m := make(PersonMap)

	for _, p := range l.persons {
		m[p.UUID] = &p.Person
	}

	return m
}

// Get returns a person
func (l *List) Get(id string) *Person {
	l.RLock()
	defer l.RUnlock()

	p := l.persons[id]

	if p == nil {
		return nil
	}
	return &p.Person
}

func (l *List) Save() error {
	configFile, err := os.Create("persons.json")
	if err != nil {
		return fmt.Errorf("persons: error saving: %s", err)
	}
	encoder := json.NewEncoder(configFile)
	encoder.SetIndent("", "\t")
	l.Lock()
	defer l.Unlock()
	err = encoder.Encode(l.persons)
	if err != nil {
		return fmt.Errorf("persons: error saving: %s", err)
	}
	return nil
}

func (l *List) Load() error {
	logrus.Info("Loading persons from json file")

	configFile, err := os.Open("persons.json")
	if err != nil {
		if os.IsNotExist(err) {

			// Add the default person, if the person list is empty
			person := PersonWithPassword{
				Person: Person{
					UUID:  uuid.New().String(),
					Name:  "J. Random",
					Email: "admin",
				},
				Password: "admin",
			}
			l.Add(&person)
			l.Save()

			logrus.Warn(err)
			return nil // We dont want to error our if the file does not exist when we start the server
		}
		return err
	}

	l.Lock()
	defer l.Unlock()
	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&l.persons); err != nil {
		return fmt.Errorf("persons: error loading: %s", err)
	}

	return nil
}
