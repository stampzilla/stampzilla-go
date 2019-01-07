package store

import (
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/logic"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/models"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/models/devices"
)

type Nodes map[string]*models.Node
type Connections map[string]*models.Connection
type UpdateCallback func(*Store) error

type Store struct {
	Nodes       Nodes
	SavedState  *logic.SavedStateStore
	Devices     *devices.List
	Connections Connections
	onUpdate    map[string][]UpdateCallback
	sync.RWMutex
}

func New() *Store {
	s := &Store{
		Nodes:       make(Nodes),
		SavedState:  logic.NewSavedStateStore(),
		Devices:     devices.NewList(),
		Connections: make(Connections),
		onUpdate:    make(map[string][]UpdateCallback, 0),
	}

	return s
}

func (store *Store) runCallbacks(area string) {
	for _, callback := range store.onUpdate[area] {
		if err := callback(store); err != nil {
			logrus.Error("store: ", err)
		}
	}
}

func (store *Store) OnUpdate(area string, callback UpdateCallback) {
	if _, ok := store.onUpdate[area]; !ok {
		store.onUpdate[area] = make([]UpdateCallback, 0)
	}
	store.Lock()
	store.onUpdate[area] = append(store.onUpdate[area], callback)
	store.Unlock()
}

// Load loads all stuff from disk.
func (store *Store) Load() error {
	// Load logic stuff
	err := store.SavedState.Load("savedstate.json")
	if err != nil {
		return err
	}

	// load all the nodes
	return store.LoadNodes()
}
