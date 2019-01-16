package main

import (
	"fmt"
	"sync"

	"github.com/vapourismo/knx-go/knx"
	"github.com/vapourismo/knx-go/knx/cemi"
	"github.com/vapourismo/knx-go/knx/dpt"
)

type config struct {
	Gateway gateway  `json:"gateway"`
	Lights  []light  `json:"lights"`
	Sensors []sensor `json:"sensors"`
	sync.Mutex
}

type gateway struct {
	Address string `json:"address"`
}

type light struct {
	ID                string `json:"id"`
	ControlSwitch     string `json:"control_switch"`
	ControlBrightness string `json:"control_brightness"`
	StateSwitch       string `json:"state_switch"`
	StateBrightness   string `json:"state_brightness"`
}

type sensor struct {
	ID          string `json:"id"`
	Motion      string `json:"motion"`
	Lux         string `json:"lux"`
	Temperature string `json:"temperature"`
}

func (light *light) Switch(tunnel *tunnel, target bool) error {
	if !tunnel.Connected {
		return fmt.Errorf("Not connected to KNX gateway")
	}
	addr, err := cemi.NewGroupAddrString(light.ControlSwitch)
	if err != nil {
		return err
	}

	cmd := knx.GroupEvent{
		Command:     knx.GroupWrite,
		Destination: addr,
		Data:        dpt.DPT_1001(target).Pack(),
	}
	return tunnel.Client.Send(cmd)
}

func (light *light) Brightness(tunnel *tunnel, target float64) error {
	if !tunnel.Connected {
		return fmt.Errorf("Not connected to KNX gateway")
	}
	addr, err := cemi.NewGroupAddrString(light.ControlBrightness)
	if err != nil {
		return err
	}

	cmd := knx.GroupEvent{
		Command:     knx.GroupWrite,
		Destination: addr,
		Data:        dpt.DPT_5001(float32(target)).Pack(),
	}
	return tunnel.Client.Send(cmd)
}
