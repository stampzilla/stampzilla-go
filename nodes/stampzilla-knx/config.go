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
	Lights  lights   `json:"lights"`
	Sensors []sensor `json:"sensors"`
	sync.Mutex
}

func (c *config) GetLight(id string) *light {
	c.Lock()
	defer c.Unlock()
	for _, v := range c.Lights {
		if v.ID == id {
			return &v
		}
	}
	return nil
}

type gateway struct {
	Address string `json:"address"`
}

type lights []light

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
	MotionTrue  string `json:"motionTrue"`
	Lux         string `json:"lux"`
	Temperature string `json:"temperature"`
	Humidity    string `json:"humidity"`
	Co2         string `json:"co2"`
	Voc         string `json:"voc"`
	AirPressure string `json:"airPressure"`
	DewPoint    string `json:"dewPoint"`
}

func (light *light) Switch(tunnel *tunnel, target bool) error {
	if !tunnel.Connected {
		return fmt.Errorf("not connected to KNX gateway")
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
		return fmt.Errorf("not connected to KNX gateway")
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
