package main

import (
	"encoding/hex"
	"log"
	"sync"

	"github.com/stamp/go-lifx/client"
	"github.com/stampzilla/stampzilla-go/protocol/devices"
)

type State struct {
	Lamps       map[string]*Lamp `json:"lamps"`
	LifxClound  StateLifxCloud   `json:"lifx_clound"`
	LanProtocol StateLanProtocol `json:"lan_protocol"`

	publishStateFunc func()

	sync.Mutex
}

type Lamp struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Ip   string `json:"ip"`

	Level float64    `json:"level"`
	Color StateColor `json:"color"`
	Power bool       `json:"power"`

	LanConnected   bool `json:"connected_lan"`
	CloudConnected bool `json:"connected_cloud"`

	Capabilities map[string]bool `json:"capabilities"`

	publishState func()
	sync.Mutex
}

type StateColor struct {
	Hue        float64 `json:"hue"`
	Saturation float64 `json:"saturation"`
	Kelvin     int     `json:"kelvin"`
}

type StateLifxCloud struct {
	State string `json:"state"`
}
type StateLanProtocol struct {
	State string `json:"state"`
}

func NewState() *State {
	return &State{Lamps: make(map[string]*Lamp), publishStateFunc: func() {}}
}

func (s *State) Lamp(id [4]byte) *Lamp {
	senderId := hex.EncodeToString(id[0:4])
	s.Lock()
	defer s.Unlock()
	if _, ok := s.Lamps[senderId]; ok {
		return s.Lamps[senderId]
	}
	return nil
}
func (s *State) AddLanDevice(light *client.Light) *Lamp {
	d := s.GetByID(light.Id())
	if d == nil {
		d = NewLamp(light.Id(), light.Label(), light.Ip.String())
		return s.Add(d)
	}

	return d
}
func (s *State) Add(d *Lamp) *Lamp {
	log.Println("Added new ", d.Id)
	d.publishState = s.publishState

	s.Lock()
	defer s.Unlock()
	defer s.publishState()

	s.Lamps[d.Id] = d
	return d
}
func (s *State) GetByID(id string) *Lamp {
	s.Lock()
	defer s.Unlock()
	if _, ok := s.Lamps[id]; ok {
		return s.Lamps[id]
	}

	return nil
}

func (s *State) RemoveDevice(id [4]byte) {
	s.Lock()
	defer s.Unlock()
	defer s.publishState()

	senderId := hex.EncodeToString(id[0:4])
	delete(s.Lamps, senderId)
}

func (s *State) publishState() {
	for _, v := range s.Lamps {
		node.Devices().Add(&devices.Device{
			Type:   "dimmableLamp",
			Name:   v.Name,
			Id:     v.Id,
			Online: v.LanConnected || v.CloudConnected,
			Node:   node.Uuid(),
			StateMap: map[string]string{
				"on":    "lamps[" + v.Id + "]" + ".power",
				"level": "lamps[" + v.Id + "]" + ".level",
			},
		})
	}

	s.publishStateFunc()
}

// -----------------------------------------------------------

func NewLamp(id, name, ip string) *Lamp {
	d := &Lamp{Name: name, Ip: ip}
	d.SetId(id)
	return d
}

func (d *Lamp) SetId(id string) {
	d.Lock()
	defer d.Unlock()
	d.Id = id

}

func (d *Lamp) SyncFromCloud(c *cloudGetAllResponse) {
	d.Level = c.Brightness * 100
	d.Color.Hue = c.Color.Hue
	d.Color.Saturation = c.Color.Saturation
	d.Color.Kelvin = c.Color.Kelvin
	d.Capabilities = c.Product.Capabilites
	d.CloudConnected = c.Connected
	d.Power = c.Power == "on"

	defer d.publishState()
}

func (d *Lamp) SyncFromLan(c *client.Light) {
	d.LanConnected = true
	d.Ip = c.Ip.String()

	defer d.publishState()
}
