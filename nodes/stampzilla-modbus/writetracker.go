package main

import (
	"sync"

	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/models/devices"
)

// writeTracker keeps polling from publishing values read before a write to the same register.
type writeTracker struct {
	sync.Mutex
	seq map[string]uint64
}

func newWriteTracker() *writeTracker {
	return &writeTracker{
		seq: make(map[string]uint64),
	}
}

// Snapshot returns the current write sequence for all registers.
func (w *writeTracker) Snapshot() map[string]uint64 {
	w.Lock()
	defer w.Unlock()
	snapshot := make(map[string]uint64, len(w.seq))
	for k, v := range w.seq {
		snapshot[k] = v
	}
	return snapshot
}

// Written marks registers as written so polls started before now will not publish them.
func (w *writeTracker) Written(names ...string) {
	w.Lock()
	defer w.Unlock()
	for _, name := range names {
		w.seq[name]++
	}
}

// Publish removes values from state that have been written since snapshot was taken and calls fn with
// the remaining state. fn is called while holding the lock so no write can be marked in between.
func (w *writeTracker) Publish(snapshot map[string]uint64, state devices.State, fn func(devices.State)) {
	w.Lock()
	defer w.Unlock()
	for name := range state {
		if w.seq[name] != snapshot[name] {
			delete(state, name)
		}
	}
	fn(state)
}
