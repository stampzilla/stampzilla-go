package main

import (
	"bytes"
	"encoding/json"
	"os"

	log "github.com/cihub/seelog"
	"github.com/facebookgo/inject"
	"github.com/koding/multiconfig"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/logic"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/metrics"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/protocol"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/websocket"
)

//TODO make config general by using a map so we can get config from ENV,file or flag.
type ServerConfig struct {
	Uuid             string
	NodePort         string `default:"8282"`
	WebPort          string `default:"8080"`
	WebRoot          string `default:"public"`
	ElasticSearch    string
	InfluxDbServer   string
	InfluxDbUser     string
	InfluxDbPassword string
}

type Startable interface {
	Start()
}

func main() {
	config := &ServerConfig{}
	m := loadMultiConfig()
	m.MustLoad(config)

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

	// Register the rest of the services
	services = append(services, &WebsocketHandler{}, config, protocol.NewNodes(), logic.NewLogic(), logic.NewScheduler(), websocket.NewRouter(), NewNodeServer(), NewWebServer())

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
