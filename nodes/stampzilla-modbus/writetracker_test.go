package main

import (
	"testing"

	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/models/devices"
	"github.com/stretchr/testify/assert"
)

func TestWriteTrackerDropsStaleValues(t *testing.T) {
	w := newWriteTracker()

	snapshot := w.Snapshot()
	polled := devices.State{"fanSpeed": 1.0, "outsideTemp": 5.0}
	w.Written("fanSpeed") // write happens while polling

	var published devices.State
	w.Publish(snapshot, polled, func(s devices.State) { published = s })
	assert.Equal(t, devices.State{"outsideTemp": 5.0}, published)

	// next poll started after the write publishes everything
	snapshot = w.Snapshot()
	polled = devices.State{"fanSpeed": 2.0, "outsideTemp": 5.0}
	w.Publish(snapshot, polled, func(s devices.State) { published = s })
	assert.Equal(t, devices.State{"fanSpeed": 2.0, "outsideTemp": 5.0}, published)
}

func TestWriteTrackerDropsPollStartedDuringWrite(t *testing.T) {
	w := newWriteTracker()

	w.Written("fanSpeed") // write started
	snapshot := w.Snapshot()
	polled := devices.State{"fanSpeed": 1.0, "outsideTemp": 5.0} // read before the serial write
	w.Written("fanSpeed")                                        // write done

	var published devices.State
	w.Publish(snapshot, polled, func(s devices.State) { published = s })
	assert.Equal(t, devices.State{"outsideTemp": 5.0}, published)
}
