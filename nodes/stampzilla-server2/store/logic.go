package store

import (
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/logic"
)

func (store *Store) GetRules() logic.Rules {
	store.Logic.RLock()
	defer store.Logic.RUnlock()
	return store.Logic.Rules
}

func (store *Store) AddOrUpdateRules(rules logic.Rules) {
	store.Logic.SetRules(rules)
	store.runCallbacks("rules")
}
