package main

import (
	"fmt"
	"time"

	volume "github.com/itchyny/volume-go"
	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models/devices"
)

func monitorVolume() {
	dev := &devices.Device{
		Name:   "Audio",
		ID:     devices.ID{ID: "audio"},
		Online: true,
		Traits: []string{"OnOff", "Volume"},
		State: devices.State{
			"on":     false,
			"volume": 0,
		},
	}
	added := false

	for {
		vol, err := volume.GetVolume()
		if err != nil {
			logrus.Errorf("get volume failed: %+v", err)
			return
		}

		mute, err := volume.GetMuted()
		if err != nil {
			logrus.Errorf("get mute failed: %+v", err)
			return
		}

		if !added {
			n.AddOrUpdate(dev)
		}

		newState := make(devices.State)
		newState["on"] = !mute
		newState["volume"] = float64(vol) / 100
		n.UpdateState(dev.ID.ID, newState)
		<-time.After(time.Second * 1)
	}
}

func commandVolume(state devices.State) error {
	if state["volume"] != nil {
		err := volume.SetVolume(int(state["volume"].(float64) * 100))
		if err != nil {
			return fmt.Errorf("set volume failed: %+v", err)
		}
	}

	if state["on"] == true {
		err := volume.Unmute()
		if err != nil {
			return fmt.Errorf("mute failed: %+v", err)
		}
	} else if state["on"] == false {
		err := volume.Mute()
		if err != nil {
			return fmt.Errorf("unmute failed: %+v", err)
		}

	}

	return nil
}
