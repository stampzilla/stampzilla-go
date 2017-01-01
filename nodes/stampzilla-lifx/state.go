package main

import (
	"encoding/hex"
	"log"
	"sync"

	"github.com/stamp/go-lifx/client"
	"github.com/stampzilla/stampzilla-go/protocol/devices"
)

// State contains the lifx node state
type State struct {
	Lamps       map[string]*Lamp `json:"lamps"`
	LifxClound  StateLifxCloud   `json:"lifx_clound"`
	LanProtocol StateLanProtocol `json:"lan_protocol"`

	publishStateFunc func()

	sync.Mutex
}

// Lamp is the state for each found lamp. Can be sourced from both lan and cloud protocols
type Lamp struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	IP   string `json:"ip"`

	Level float64    `json:"level"`
	Color StateColor `json:"color"`
	Power bool       `json:"power"`

	LanConnected   bool `json:"connected_lan"`
	CloudConnected bool `json:"connected_cloud"`

	Capabilities map[string]bool `json:"capabilities"`

	publishState func()
	sync.Mutex
}

// StateColor describes the current color at the lifx lamp
type StateColor struct {
	Hue        float64 `json:"hue"`
	Saturation float64 `json:"saturation"`
	Kelvin     int     `json:"kelvin"`
}

// StateLifxCloud is the current state of the Lifx cloud poller
type StateLifxCloud struct {
	State         string `json:"state"`
	FoundDevices  int    `json:"found_devices"`
	OnlineDevices int    `json:"online_devices"`
}

// StateLanProtocol is the current state of the Lifx lan protocol
type StateLanProtocol struct {
	State         string `json:"state"`
	FoundDevices  int    `json:"found_devices"`
	OnlineDevices int    `json:"online_devices"`
}

// NewState creates a new State struct and initializes all values
func NewState() *State {
	return &State{Lamps: make(map[string]*Lamp), publishStateFunc: func() {}}
}

// Lamp Fetches a lamp from the state list based on id [4]byte
func (s *State) Lamp(id [4]byte) *Lamp {
	senderID := hex.EncodeToString(id[0:4])
	s.Lock()
	defer s.Unlock()
	if _, ok := s.Lamps[senderID]; ok {
		return s.Lamps[senderID]
	}
	return nil
}

// GetByID fetches a lamp from the state list based on id string
func (s *State) GetByID(id string) *Lamp {
	s.Lock()
	defer s.Unlock()
	if _, ok := s.Lamps[id]; ok {
		return s.Lamps[id]
	}

	return nil
}

// AddLanDevice checks if the device already exists, else adds a new one based on the information provided
func (s *State) AddLanDevice(light *client.Light) *Lamp {
	d := s.GetByID(light.Id())
	if d == nil {
		d = NewLamp(light.Id(), light.Label(), light.Ip.String())
		return s.Add(d)
	}

	return d
}

// Add a Lamp struct to the lamps list
func (s *State) Add(d *Lamp) *Lamp {
	log.Println("Added new ", d.ID)
	d.publishState = s.publishState

	s.Lock()
	defer s.Unlock()
	defer s.publishState()

	s.Lamps[d.ID] = d
	return d
}

// RemoveDevice removes a lamp that are no longer available from the list
func (s *State) RemoveDevice(id [4]byte) {
	s.Lock()
	defer s.Unlock()
	defer s.publishState()

	senderID := hex.EncodeToString(id[0:4])
	delete(s.Lamps, senderID)
}

func (s *State) publishState() {
	for _, v := range s.Lamps {
		node.Devices().Add(&devices.Device{
			Type:   "dimmableLamp",
			Name:   v.Name,
			Id:     v.ID,
			Online: v.LanConnected || v.CloudConnected,
			StateMap: map[string]string{
				"on":    "lamps[" + v.ID + "]" + ".power",
				"level": "lamps[" + v.ID + "]" + ".level",
			},
		})
	}

	s.publishStateFunc()
}

// -----------------------------------------------------------

// NewLamp initiates a new Lamp type
func NewLamp(id, name, ip string) *Lamp {
	d := &Lamp{Name: name, IP: ip}
	d.SetID(id)
	return d
}

// SetID setter for the idf field
func (d *Lamp) SetID(id string) {
	d.Lock()
	defer d.Unlock()
	d.ID = id

}

// SyncFromCloud syncronizes data from a cloudGetAllReponse struct received from the lifx cloud
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

// SyncFromLan syncronizes data from the lan protocol
func (d *Lamp) SyncFromLan(c *client.Light) {
	d.LanConnected = true
	d.IP = c.Ip.String()

	defer d.publishState()
}
