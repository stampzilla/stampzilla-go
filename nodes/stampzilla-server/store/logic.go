package store

import (
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/logic"
)

func (store *Store) GetRules() logic.Rules {
	store.Logic.RLock()
	defer store.Logic.RUnlock()
	return store.Logic.Rules
}

func (store *Store) AddOrUpdateRules(rules logic.Rules) {
	store.Logic.SetRules(rules)
	store.Logic.Save()
	store.runCallbacks("rules")
}

func (store *Store) GetSavedStates() logic.SavedStates {
	return store.Logic.StateStore.All()
}

func (store *Store) AddOrUpdateSavedStates(s logic.SavedStates) {
	store.SavedState.SetState(s)
	store.SavedState.Save()
	store.runCallbacks("savedstates")
}

func (store *Store) GetScheduledTasks() logic.Tasks {
	return store.Scheduler.Tasks()
}

func (store *Store) AddOrUpdateScheduledTasks(tasks logic.Tasks) {
	store.Scheduler.SetTasks(tasks)
	store.Scheduler.Save()
	store.runCallbacks("scheduledtasks")
}
