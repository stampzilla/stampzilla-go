package notifications

import (
	"bytes"
	"encoding/json"
	"os"

	log "github.com/cihub/seelog"
	"github.com/koding/multiconfig"
)

type Router struct {
	Config RouterConfig

	transports map[string][]Transport
}

type RouterConfig struct {
	Transports map[string]interface{}
	Routes     map[string][]string
}

func NewRouter() *Router {
	return &Router{
		transports: make(map[string][]Transport),
		Config: RouterConfig{
			Transports: make(map[string]interface{}),
			Routes:     make(map[string][]string),
		},
	}
}

type Transport interface {
	Start()
	Dispatch(note Notification)
	Stop()
}

func (self *Router) Load(configFileName string) error {
	m := multiconfig.NewWithPath(configFileName)
	err := m.Load(self.Config)
	if err != nil {
		log.Error("Failed to read config file (", configFileName, ")", err)
	}
	return err
}

func (self *Router) Save(configFileName string) error {
	configFile, err := os.Create(configFileName)
	if err != nil {
		log.Error("Failed to create config file (", configFileName, ")", err.Error())
		return err
	}

	log.Info("Save config: ", self.Config)
	var out bytes.Buffer
	b, err := json.MarshalIndent(self.Config, "", "\t")
	if err != nil {
		log.Error("Failed to marshal json", err)
		return err
	}
	json.Indent(&out, b, "", "\t")
	out.WriteTo(configFile)

	return nil
}

func (self *Router) Start() {
	self.Load("notifications.json")
	//self.Config.Transports["smtp"] = &Smtp{}
	self.Save("notifications.json")
}

func (self *Router) AddTransport(t interface{}, levels []string) {
	if transport, ok := t.(Transport); ok {
		for _, level := range levels {
			log.Infof("Notifications - added transport (%T) for level %s", transport, level)
			self.transports[level] = append(self.transports[level], transport)
		}
		transport.Start()

		return
	}

	log.Warnf("Notifications - Added transport (%T) do not fullfill the transport interface", t)
}

func (self *Router) Dispatch(msg Notification) {
	if transports, ok := self.transports[msg.Level]; ok {
		for _, t := range transports {
			t.Dispatch(msg)
		}
	}
}
