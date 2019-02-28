package store

import (
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/logic"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models/devices"
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
	Server       map[string]map[string]devices.State

	onUpdate []UpdateCallback
	sync.RWMutex
}

func New(l *logic.Logic, s *logic.Scheduler, sss *logic.SavedStateStore) *Store {
	store := &Store{
		Nodes:        make(Nodes),
		SavedState:   sss,
		Devices:      devices.NewList(),
		Connections:  make(Connections),
		Logic:        l,
		Scheduler:    s,
		Certificates: make([]Certificate, 0),
		Requests:     make([]Request, 0),
		Server:       make(map[string]map[string]devices.State),
	}

	l.OnReportState(func(uuid string, state devices.State) {
		store.AddOrUpdateServer("rules", uuid, state)
	})

	return store
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
	if err := store.SavedState.Load(); err != nil {
		return err
	}
	if err := store.Logic.Load(); err != nil {
		return err
	}

	if err := store.Scheduler.Load(); err != nil {
		return err
	}

	// load all the nodes
	return store.LoadNodes()
}
