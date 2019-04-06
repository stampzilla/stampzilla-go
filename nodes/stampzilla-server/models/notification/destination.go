package notification

import (
	"sync"

	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models"
)

type Destination struct {
	ID           string        `json:"id"`
	Type         string        `json:"type"`
	Labels       models.Labels `json:"labels"`
	Sender       Sender        `json:"sender"`
	Destinations []string      `json:"destinations"`
}

func (d *Destination) Equal(dest *Destination) bool {
	if d.ID != dest.ID {
		return false
	}
	if d.Type != dest.Type {
		return false
	}
	// TODO: compare!
	//if d.Labels != dest.Labels {
	//return false
	//}
	if d.Sender != dest.Sender {
		return false
	}
	// TODO: compare!
	//if d.Destinations != dest.Destinations {
	//return false
	//}

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
	d.destinations[dest.ID] = dest
	d.Unlock()
}

func (d *Destinations) Get(id string) *Destination {
	d.RLock()
	defer d.RUnlock()
	return d.destinations[id]
}

func (d *Destinations) All() map[string]*Destination {
	d.RLock()
	defer d.RUnlock()
	return d.destinations
}

func (d *Destinations) Remove(id string) {
	d.Lock()
	delete(d.destinations, id)
	d.Unlock()
}
