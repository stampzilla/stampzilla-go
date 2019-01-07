package interfaces

import "github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/models/devices"

// StateSyncer takes a map of node uuids and their desired DeviceState
type StateSyncer interface {
	SyncState(map[string]devices.State)
}
