package store

import (
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/logic"
	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/models"
	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/models/devices"
	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/models/notification"
	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/models/persons"
)

type (
	Nodes              map[string]*models.Node
	Connections        map[string]*models.Connection
	UpdateCallback     func(string, *Store) error
	UserDemoteCallback func(string) error
)

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
	Persons      persons.List

	Destinations *notification.Destinations
	Senders      *notification.Senders

	onUpdate     []UpdateCallback
	onUserDemote []UserDemoteCallback
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
		Destinations: notification.NewDestinations(),
		Senders:      notification.NewSenders(),
		Persons:      persons.NewList(),
	}

	l.OnReportState(func(uuid string, state devices.State) {
		store.AddOrUpdateServer("rules", uuid, state)
	})

	l.OnTriggerDestination(store.TriggerDestination)

	return store
}

func (store *Store) runCallbacks(area string) {
	for _, callback := range store.onUpdate {
		if err := callback(area, store); err != nil {
			logrus.Error("store callback: ", err)
		}
	}
}

func (store *Store) OnUpdate(callback UpdateCallback) {
	store.Lock()
	store.onUpdate = append(store.onUpdate, callback)
	store.Unlock()
}

func (store *Store) OnUserDemote(callback UserDemoteCallback) {
	store.Lock()
	store.onUserDemote = append(store.onUserDemote, callback)
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

	if err := store.Destinations.Load("destinations.json"); err != nil {
		return err
	}

	if err := store.Senders.Load("senders.json"); err != nil {
		return err
	}

	if err := store.Persons.Load(); err != nil {
		return err
	}

	// load all the nodes
	return store.LoadNodes()
}
