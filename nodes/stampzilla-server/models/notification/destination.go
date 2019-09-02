package notification

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models"
)

type Destination struct {
	UUID         string        `json:"uuid"`
	Name         string        `json:"name"`
	Type         string        `json:"type"`
	Labels       models.Labels `json:"labels"`
	Sender       string        `json:"sender"`
	Destinations []string      `json:"destinations"`
}

func (d *Destination) Equal(dest *Destination) bool {
	if d.UUID != dest.UUID {
		return false
	}
	if d.Type != dest.Type {
		return false
	}
	if d.Name != dest.Name {
		return false
	}
	// TODO: compare!
	//if d.Labels != dest.Labels {
	//return false
	//}
	if d.Sender != dest.Sender {
		return false
	}
	if !EqualStringMap(d.Destinations, dest.Destinations) {
		return false
	}

	return true
}

func EqualStringMap(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

type Destinations struct {
	destinations map[string]*Destination
	sync.RWMutex
}

func NewDestinations() *Destinations {
	return &Destinations{
		destinations: make(map[string]*Destination),
	}
}

func (d *Destinations) Add(dest *Destination) {
	d.Lock()
	d.destinations[dest.UUID] = dest
	d.Unlock()
}

func (d *Destinations) Get(uuid string) *Destination {
	d.RLock()
	defer d.RUnlock()
	return d.destinations[uuid]
}

func (d *Destinations) Trigger(uuid, body string) error {
	d.RLock()
	defer d.RUnlock()

	return fmt.Errorf("Not implemented yet")
}

func (d *Destinations) All() map[string]*Destination {
	d.RLock()
	defer d.RUnlock()
	return d.destinations
}

func (d *Destinations) Remove(uuid string) {
	d.Lock()
	delete(d.destinations, uuid)
	d.Unlock()
}

// Save saves the rules to rules.json
func (s *Destinations) Save() error {
	configFile, err := os.Create("destinations.json")
	if err != nil {
		return fmt.Errorf("destinations: error saving destinations: %s", err.Error())
	}
	encoder := json.NewEncoder(configFile)
	encoder.SetIndent("", "\t")
	s.Lock()
	defer s.Unlock()
	err = encoder.Encode(s.destinations)
	if err != nil {
		return fmt.Errorf("destinations: error saving destinations: %s", err.Error())
	}
	return nil
}

//Load loads the rules from rules.json
func (s *Destinations) Load() error {
	logrus.Debug("destinations: loading rules from destinations.json")
	configFile, err := os.Open("destinations.json")
	if err != nil {
		if os.IsNotExist(err) {
			logrus.Warn(err)
			return nil // We dont want to error our if the file does not exist when we start the server
		}
		return fmt.Errorf("destinations: error loading destinations.json: %s", err.Error())
	}

	s.Lock()
	defer s.Unlock()
	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&s.destinations); err != nil {
		return fmt.Errorf("logic: error loading rules.json: %s", err.Error())
	}

	//TODO loop over rules and generate UUIDs if needed. If it was needed save the rules again

	return nil
}
