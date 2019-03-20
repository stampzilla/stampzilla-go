package main

import (
	"time"

	sigar "github.com/cloudfoundry/gosigar"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models/devices"
)

func monitorHealth() {
	dev := &devices.Device{
		Name:   "Health",
		ID:     devices.ID{ID: "health"},
		Online: true,
		State:  devices.State{},
	}
	n.AddOrUpdate(dev)

	for {
		uptime := sigar.Uptime{}
		uptime.Get()
		avg := sigar.LoadAverage{}
		avg.Get()

		newState := make(devices.State)
		newState["uptime"] = uptime.Format()
		newState["load_1"] = avg.One
		newState["load_5"] = avg.Five
		newState["load_15"] = avg.Fifteen
		n.UpdateState(dev.ID.ID, newState)

		<-time.After(time.Second * 1)
	}
}
