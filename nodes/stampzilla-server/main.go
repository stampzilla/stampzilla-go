package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	log "github.com/cihub/seelog"
	"github.com/facebookgo/inject"
	"github.com/koding/multiconfig"
	"github.com/pborman/uuid"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/logic"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/metrics"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/notifications"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/protocol"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/websocket"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/websocket/handlers"
	"github.com/stampzilla/stampzilla-go/pkg/notifier"
)

var VERSION string = "dev"
var BUILD_DATE string = ""

//TODO make config general by using a map so we can get config from ENV,file or flag.
type ServerConfig struct {
	Uuid             string
	NodePort         string `default:"8282"`
	WebPort          string `default:"8080"`
	WebRoot          string `default:"public/dist"`
	ElasticSearch    string
	InfluxDbServer   string
	InfluxDbUser     string
	InfluxDbPassword string
}

type Startable interface {
	Start()
}

var notify *notifier.Notify

func main() {
	printVersion := flag.Bool("v", false, "Prints current version")
	flag.Parse()

	if *printVersion != false {
		fmt.Println(VERSION + " (" + BUILD_DATE + ")")
		os.Exit(0)
	}

	config := &ServerConfig{}
	m := loadMultiConfig()
	m.MustLoad(config)

	// Create an uuid
	if config.Uuid == "" {
		config.Uuid = uuid.New()
	}

	// Load logger
	logger, err := log.LoggerFromConfigAsFile("logconfig.xml")
	if err != nil {
		testConfig := `
			<seelog type="sync" asyncinterval="1000" minlevel="trace">
				<outputs>
					<filter levels="trace">
						<console formatid="colored-trace"/>
					</filter>
					<filter levels="debug">
						<console formatid="colored-debug"/>
					</filter>
					<filter levels="info">
						<console formatid="colored-info"/>
					</filter>
					<filter levels="warn">
						<console formatid="colored-warn"/>
					</filter>
					<filter levels="error">
						<console formatid="colored-error"/>
					</filter>
					<filter levels="critical">
						<console formatid="colored-critical"/>
					</filter>
				</outputs>
				<formats>
					<format id="colored-trace"  format="%Date %Time %EscM(40)%Level%EscM(49) - %File:%Line - %Msg%n%EscM(0)"/>
					<format id="colored-debug"  format="%Date %Time %EscM(45)%Level%EscM(49) - %File:%Line - %Msg%n%EscM(0)"/>
					<format id="colored-info"  format="%Date %Time %EscM(46)%Level%EscM(49) - %File:%Line - %Msg%n%EscM(0)"/>
					<format id="colored-warn"  format="%Date %Time %EscM(43)%Level%EscM(49) - %File:%Line - %Msg%n%EscM(0)"/>
					<format id="colored-error"  format="%Date %Time %EscM(41)%Level%EscM(49) - %File:%Line - %Msg%n%EscM(0)"/>
					<format id="colored-critical"  format="%Date %Time %EscM(41)%Level%EscM(49) - %File:%Line - %Msg%n%EscM(0)"/>
				</formats>
			</seelog>`
		logger, _ = log.LoggerFromConfigAsBytes([]byte(testConfig))
	}
	log.ReplaceLogger(logger)

	services := make([]interface{}, 0)

	// Register metrics loggers
	if config.ElasticSearch != "" {
		log.Info("Starting ElasticSearch metrics logger")
		es := NewElasticSearch()
		services = append(services, es)
	}
	if config.InfluxDbServer != "" {
		i := NewInfluxDb()
		services = append(services, i)
	}

	notificationRouter := notifications.NewRouter()
	notificationRouter.Uuid = config.Uuid
	notificationRouter.Name = "server"
	notify = notifier.New(notificationRouter)

	params := map[string]string{
		"version":    VERSION,
		"build_date": BUILD_DATE,
	}
	msg, err := json.Marshal(params)
	wsr := websocket.NewRouter()
	wsr.AddClientConnectHandler(func() *websocket.Message {
		return &websocket.Message{Type: "parameters", Data: msg}
	})

	// Register the rest of the services
	services = append(
		services,
		logic.NewActionService(),
		&handlers.Nodes{},
		&handlers.Actions{},
		&handlers.Rules{},
		&handlers.Schedule{},
		&handlers.Devices{},
		config,
		protocol.NewNodes(),
		protocol.NewDevices(),
		logic.NewLogic(),
		logic.NewScheduler(),
		wsr,
		NewNodeServer(),
		NewWebServer(),
		notificationRouter,
	)

	//Add metrics service if we have any loggers (Elasticsearch, influxdb, graphite etc)
	if loggers := getLoggers(services); len(loggers) != 0 {
		m := metrics.New()
		log.Info("Detected metrics loggers, starting up")
		for _, l := range loggers {
			m.AddLogger(l)
		}
		services = append(services, m)
	}

	err = inject.Populate(services...)
	if err != nil {
		panic(err)
	}

	saveConfigToFile(config)

	StartServices(services)

	notify.Info("Server started and ready")
	select {}
}

func StartServices(services []interface{}) {
	for _, s := range services {
		if s, ok := s.(Startable); ok {
			s.Start()
		}
	}
}

func getLoggers(services []interface{}) []metrics.Logger {
	var loggers []metrics.Logger
	for _, s := range services {
		if s, ok := s.(metrics.Logger); ok {
			loggers = append(loggers, s)
		}
	}
	return loggers
}

func saveConfigToFile(config *ServerConfig) {
	configFile, err := os.Create("config.json")
	if err != nil {
		log.Error("creating config file", err.Error())
	}

	log.Info("Save config: ", config)
	var out bytes.Buffer
	b, err := json.MarshalIndent(config, "", "\t")
	if err != nil {
		log.Error("error marshal json", err)
	}
	json.Indent(&out, b, "", "\t")
	out.WriteTo(configFile)
}

func loadMultiConfig() *multiconfig.DefaultLoader {
	loaders := []multiconfig.Loader{}

	// Read default values defined via tag fields "default"
	loaders = append(loaders, &multiconfig.TagLoader{})

	if _, err := os.Stat("config.json"); err == nil {
		loaders = append(loaders, &multiconfig.JSONLoader{Path: "config.json"})
	}

	e := &multiconfig.EnvironmentLoader{}
	e.Prefix = "STAMPZILLA"
	f := &multiconfig.FlagLoader{}
	f.EnvPrefix = "STAMPZILLA"

	loaders = append(loaders, e, f)
	loader := multiconfig.MultiLoader(loaders...)

	d := &multiconfig.DefaultLoader{}
	d.Loader = loader
	d.Validator = multiconfig.MultiValidator(&multiconfig.RequiredValidator{})
	return d

}
