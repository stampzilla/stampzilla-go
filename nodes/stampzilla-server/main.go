package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"

	log "github.com/cihub/seelog"
	"github.com/facebookgo/inject"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/logic"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/metrics"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/protocol"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/websocket"
)

//TODO make config general by using a map so we can get config from ENV,file or flag.
type ServerConfig struct {
	NodePort         string
	WebPort          string
	WebRoot          string
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

	flag.StringVar(&config.NodePort, "node-port", "8282", "Stampzilla NodeServer port")
	flag.StringVar(&config.WebPort, "web-port", "8080", "Webserver port")
	flag.StringVar(&config.WebRoot, "web-root", "public", "Webserver root")
	flag.StringVar(&config.ElasticSearch, "elasticsearch", "", "Address to an ElasticSearch host. Ex: http://hostname:9200/test/test")
	flag.StringVar(&config.InfluxDbServer, "influxdbserver", "", "Address to an InfluxDb host. Ex: http://localhost:8086")
	flag.StringVar(&config.InfluxDbUser, "influxdbuser", "", "InfluxDb user. ")
	flag.StringVar(&config.InfluxDbPassword, "influxdbpassword", "", "InfluxDb password. ")
	flag.Parse()

	readConfigFromFile("config.json", config)
	getConfigFromEnv(config)

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

func getConfigFromEnv(config *ServerConfig) {

	//TODO make prettier and generate from map with both ENV and flags
	if val := os.Getenv("STAMPZILLA_WEBROOT"); val != "" {
		config.WebRoot = val
	}
}

func readConfigFromFile(fn string, config *ServerConfig) {
	configFile, err := os.Open(fn)
	if err != nil {
		log.Error("opening config file", err.Error())
		return
	}

	newConfig := &ServerConfig{}
	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&newConfig); err != nil {
		log.Error("parsing config file", err.Error())
	}

	//Command line arguments has higher priority. Only implemented for config.InfluxDbServer yet
	//TODO generalize using reflect to itearate over config struct so check all
	if config.InfluxDbServer != "" {
		log.Info("config.InfluxDbServer != \"\"")
		newConfig.InfluxDbServer = config.InfluxDbServer
	}

	*config = *newConfig
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
