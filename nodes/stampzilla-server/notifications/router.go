package notifications

import (
	"bytes"
	"encoding/json"
	"os"

	log "github.com/cihub/seelog"
	"github.com/koding/multiconfig"
)

type Router struct {
	Config *RouterConfig

	transports map[NotificationLevel][]Transport
}

type RouterConfig struct {
	Transports map[string]interface{}
	Routes     map[string][]string
}

func NewRouter() *Router {
	return &Router{
		transports: make(map[NotificationLevel][]Transport),
		Config: &RouterConfig{
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
	log.Info("Load notifications config: ", configFileName)
	m := multiconfig.NewWithPath(configFileName)
	err := m.Load(self.Config)
	if err != nil {
		log.Error("Failed to read config file (", configFileName, ")", err)
	}

	for transport, config := range self.Config.Transports {
		var t interface{}

		// Create an instance of the transport
		switch transport {
		case "Smtp":
			t = &Smtp{}
		case "Exec":
			t = &Exec{}
		default:
			log.Errorf("Failed to create instance of transport \"%s\", no such transport is defined", transport)
		}

		if t != nil {
			// Add the config
			// TODO: Maybe find a better solution to load the config here.. // stamp
			configEncoded, _ := json.Marshal(config)
			json.Unmarshal(configEncoded, t)

			// Replace the loaded map with the real instance instead
			self.Config.Transports[transport] = t

			// Add the routes if there is any defined
			if levels, ok := self.Config.Routes[transport]; ok {
				// Register the new transport
				self.AddTransport(t, levels)
			}
		}
	}
	return err
}

func (self *Router) Save(configFileName string) error {
	log.Info("Save notifications config: ", configFileName)

	configFile, err := os.Create(configFileName)
	if err != nil {
		log.Error("Failed to create config file (", configFileName, ")", err.Error())
		return err
	}

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
	err := self.Load("notifications.json")

	// Dont resave if we failed to load the file
	if err == nil {
		self.Save("notifications.json")
	}
}

func (self *Router) AddTransport(t interface{}, levels []string) {
	if transport, ok := t.(Transport); ok {
		for _, level := range levels {
			l := NewNotificationLevel(level)
			log.Infof("Notifications - added transport (%T) for level %s", transport, l)
			self.transports[l] = append(self.transports[l], transport)
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
		return
	}

	log.Warnf("Notification type \"%s\" dropped, no one is listening", msg.Level)
}
