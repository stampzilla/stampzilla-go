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
type UpdateCallback func(string, *Store) error

type Store struct {
	Nodes        Nodes
	SavedState   *logic.SavedStateStore
	Logic        *logic.Logic
	Scheduler    *logic.Scheduler
	Devices      *devices.List
	Connections  Connections
	Certificates []Certificate
	Requests     []Request

	onUpdate []UpdateCallback
	sync.RWMutex
}

func New(l *logic.Logic, s *logic.Scheduler, sss *logic.SavedStateStore) *Store {
	return &Store{
		Nodes:        make(Nodes),
		SavedState:   sss,
		Devices:      devices.NewList(),
		Connections:  make(Connections),
		Logic:        l,
		Scheduler:    s,
		Certificates: make([]Certificate, 0),
		Requests:     make([]Request, 0),
	}
}

func (store *Store) runCallbacks(area string) {
	for _, callback := range store.onUpdate {
		if err := callback(area, store); err != nil {
			logrus.Error("store: ", err)
		}
	}
}

func (store *Store) OnUpdate(callback UpdateCallback) {
	store.Lock()
	store.onUpdate = append(store.onUpdate, callback)
	store.Unlock()
}

// Load loads all stuff from disk.
func (store *Store) Load() error {
	// Load logic stuff
	err := store.SavedState.Load("savedstate.json")
	if err != nil {
		return err
	}
	err = store.Logic.Load("rules.json")
	if err != nil {
		return err
	}

	err = store.Scheduler.Load()
	if err != nil {
		return err
	}

	// load all the nodes
	return store.LoadNodes()
}
