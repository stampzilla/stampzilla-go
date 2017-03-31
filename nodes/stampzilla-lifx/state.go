package main

import (
	"encoding/hex"
	"encoding/json"
	"log"
	"sync"

	"github.com/stamp/go-lifx/client"
	"github.com/stampzilla/stampzilla-go/protocol/devices"
)

// State contains the lifx node state
type State struct {
	lamps       map[string]Lamp  `json:"lamps"`
	lifxCloud   StateLifxCloud   `json:"lifx_clound"`
	lanProtocol StateLanProtocol `json:"lan_protocol"`

	publishStateFunc func()

	sync.RWMutex
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
	sync.RWMutex
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
	return &State{lamps: make(map[string]Lamp), publishStateFunc: func() {}}
}

// Lamp Fetches a lamp from the state list based on id [4]byte
func (s *State) Lamp(id [4]byte) *Lamp {
	senderID := hex.EncodeToString(id[0:4])
	s.RLock()
	defer s.RUnlock()

	if _, ok := s.lamps[senderID]; ok {
		l := s.lamps[senderID]
		return &l
	}
	return nil
}

// GetByID fetches a lamp from the state list based on id string
func (s *State) GetByID(id string) *Lamp {
	s.RLock()
	defer s.RUnlock()

	if _, ok := s.lamps[id]; ok {
		l := s.lamps[id]
		return &l
	}

	return nil
}

// AddLanDevice checks if the device already exists, else adds a new one based on the information provided
func (s *State) AddLanDevice(light *client.Light) *Lamp {
	d := s.GetByID(light.Id())
	if d == nil {
		d = NewLamp(light.Id(), light.Label(), light.Ip.String())
		d.LanConnected = true
		return s.Add(d)
	}

	d.LanConnected = true

	return d
}

// Add a Lamp struct to the lamps list
func (s *State) Add(d *Lamp) *Lamp {
	log.Println("Added new ", d.ID)
	d.publishState = s.publishState

	s.Lock()
	s.lamps[d.ID] = *d
	s.Unlock()

	s.publishState()

	return d
}

// RemoveDevice removes a lamp that are no longer available from the list
func (s *State) RemoveDevice(id [4]byte) {
	senderID := hex.EncodeToString(id[0:4])

	s.Lock()
	delete(s.lamps, senderID)
	s.Unlock()

	s.publishState()
}

func (s *State) publishState() {
	s.RLock()
	for _, v := range s.lamps {
		v.RLock()
		d := &devices.Device{
			Type:   "dimmableLamp",
			Name:   v.Name,
			Id:     v.ID,
			Online: v.LanConnected || v.CloudConnected,
			StateMap: map[string]string{
				"on":    "lamps[" + v.ID + "]" + ".power",
				"level": "lamps[" + v.ID + "]" + ".level",
			},
		}
		v.RUnlock()

		node.Devices().Add(d)
	}
	s.RUnlock()

	s.publishStateFunc()
}

func (s *State) MarshalJSON() ([]byte, error) {
	type alias struct {
		Lamps       map[string]Lamp  `json:"lamps"`
		LifxCloud   StateLifxCloud   `json:"lifx_clound"`
		LanProtocol StateLanProtocol `json:"lan_protocol"`
	}

	s.RLock()
	a := alias{
		Lamps:       make(map[string]Lamp),
		LifxCloud:   s.lifxCloud,
		LanProtocol: s.lanProtocol,
	}
	for k, v := range s.lamps {
		v.Lock()
		a.Lamps[k] = v
		v.Unlock()
	}
	defer s.RUnlock()

	return json.Marshal(a)
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
	d.ID = id
	d.Unlock()
}

// SyncFromCloud syncronizes data from a cloudGetAllReponse struct received from the lifx cloud
func (d *Lamp) SyncFromCloud(c *cloudGetAllResponse) {
	d.Lock()
	d.Level = c.Brightness * 100
	d.Color.Hue = c.Color.Hue
	d.Color.Saturation = c.Color.Saturation
	d.Color.Kelvin = c.Color.Kelvin
	d.Capabilities = c.Product.Capabilites
	d.CloudConnected = c.Connected
	d.Power = c.Power == "on"
	d.Unlock()

	defer d.publishState()
}

// SyncFromLan syncronizes data from the lan protocol
func (d *Lamp) SyncFromLan(c *client.Light) {
	d.Lock()
	d.LanConnected = true
	d.IP = c.Ip.String()
	d.Unlock()

	defer d.publishState()
}
