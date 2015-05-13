package main

import (
	"flag"
	"os"

	log "github.com/cihub/seelog"
	"github.com/facebookgo/inject"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/logic"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/protocol"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/websocket"
)

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

	if val := os.Getenv("STAMPZILLA_WEBROOT"); val != "" {
		config.WebRoot = val
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

	nodes := protocol.NewNodes()
	scheduler := logic.NewScheduler()
	logic := logic.NewLogic()
	//scheduler.CreateExampleFile()
	//return
	webServer := NewWebServer()
	nodeServer := NewNodeServer()
	wsrouter := websocket.NewRouter()
	wsHandler := &WebsocketHandler{}

	if config.ElasticSearch != "" {
		es := NewElasticSearch()
		services = append(services, es)
	}
	if config.InfluxDbServer != "" {
		i := NewInfluxDb()
		services = append(services, i)
	}

	//Note. webServer must be started last since its blocking.
	services = append(services, config, nodes, logic, scheduler, nodeServer, wsrouter, wsHandler, webServer)
	err = inject.Populate(services...)
	if err != nil {
		panic(err)
	}

	for _, s := range services {
		if s, ok := s.(Startable); ok {
			s.Start()
		}
	}
}
