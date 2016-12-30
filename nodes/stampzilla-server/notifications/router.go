package notifications

import (
	"bytes"
	"encoding/json"
	"os"

	log "github.com/cihub/seelog"
)

type Router interface {
	Dispatch(msg Notification)
	Send(data interface{})
}

type router struct {
	Uuid   string
	Name   string
	Config *routerConfig

	transports map[NotificationLevel][]Transport
}

type routerConfig struct {
	Transports map[string]interface{}
	Routes     map[string][]string
}

func NewRouter() *router {
	return &router{
		transports: make(map[NotificationLevel][]Transport),
		Config: &routerConfig{
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

func (self *router) Load(configFileName string) error {
	log.Info("Load notifications config: ", configFileName)
	file, err := os.Open(configFileName)
	if err != nil {
		log.Error(err)
	}
	defer file.Close()

	jsonParser := json.NewDecoder(file)
	err = jsonParser.Decode(self.Config)
	if err != nil {
		log.Error("Failed to read config file (", configFileName, ") ", err)
		return err
	}

	for transport, config := range self.Config.Transports {
		var t Transport

		// Create an instance of the transport
		switch transport {
		case "Smtp":
			t = &Smtp{}
		case "Exec":
			t = &Exec{}
		case "Nma":
			t = &Nma{}
		default:
			log.Errorf("Failed to create instance of transport \"%s\", no such transport is defined", transport)
		}

		//  TODO: USE Unmarshaler? (jonaz) <Fri 21 Aug 2015 10:23:51 PM CEST>

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

func (self *router) Save(configFileName string) error {
	log.Info("Save notifications config: ", configFileName)

	configFile, err := os.Create(configFileName)
	if err != nil {
		log.Error("Failed to create config file (", configFileName, ") ", err.Error())
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

func (self *router) Start() {
	err := self.Load("notifications.json")

	// Dont resave if we failed to load the file
	if err == nil {
		self.Save("notifications.json")
	}
}

func (self *router) AddTransport(transport Transport, levels []string) {
	for _, level := range levels {
		l := NewNotificationLevel(level)
		//log.Infof("Notifications - added transport (%T) for level %s", transport, l)
		self.transports[l] = append(self.transports[l], transport)
	}
	transport.Start()

	return

}

func (self *router) Dispatch(msg Notification) {
	if transports, ok := self.transports[msg.Level]; ok {
		for _, t := range transports {
			//log.Infof("Dispatching notification type \"%s\" to \"%T\"", msg.Level, t)
			t.Dispatch(msg)
		}
		return
	}

	log.Warnf("Notification type \"%s\" dropped, no one is listening", msg.Level)
}

func (self *router) Send(data interface{}) {
	if n, ok := data.(Notification); ok {
		n.SourceUuid = self.Uuid
		n.Source = self.Name
		self.Dispatch(n)
	}
}

//func (self *Router) Critical(message string) {
//self.Dispatch(Notification{
//Source:     self.Name,
//SourceUuid: self.Uuid,
//Level:      CriticalLevel,
//Message:    message,
//})
//}

//func (self *Router) Error(message string) {
//self.Dispatch(Notification{
//Source:     self.Name,
//SourceUuid: self.Uuid,
//Level:      ErrorLevel,
//Message:    message,
//})
//}

//func (self *Router) Warn(message string) {
//self.Dispatch(Notification{
//Source:     self.Name,
//SourceUuid: self.Uuid,
//Level:      WarnLevel,
//Message:    message,
//})
//}

//func (self *Router) Info(message string) {
//self.Dispatch(Notification{
//Source:     self.Name,
//SourceUuid: self.Uuid,
//Level:      InfoLevel,
//Message:    message,
//})
//}
