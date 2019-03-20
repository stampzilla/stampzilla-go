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

	for {
		uptime := sigar.Uptime{}
		uptime.Get()
		avg := sigar.LoadAverage{}
		avg.Get()

		dev.State["uptime"] = uptime.Format()
		dev.State["load_1"] = avg.One
		dev.State["load_5"] = avg.Five
		dev.State["load_15"] = avg.Fifteen

		n.AddOrUpdate(dev)
		<-time.After(time.Second * 1)
	}
}
