package main

import (
	"flag"
	"os"

	log "github.com/cihub/seelog"
	"github.com/facebookgo/inject"
	"github.com/stampzilla/stampzilla-go/stampzilla-server/logic"
	"github.com/stampzilla/stampzilla-go/stampzilla-server/protocol"
	"github.com/stampzilla/stampzilla-go/stampzilla-server/websocket"
)

type ServerConfig struct {
	NodePort string
	WebPort  string
	WebRoot  string
}

func main() {

	config := &ServerConfig{}

	flag.StringVar(&config.NodePort, "node-port", "8282", "Stampzilla NodeServer port")
	flag.StringVar(&config.WebPort, "web-port", "8080", "Webserver port")
	flag.StringVar(&config.WebRoot, "web-root", "public", "Webserver root")
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
					<format id="colored-trace"  format="%Time %EscM(40)%Level%EscM(49) - %File:%Line - %Msg%n%EscM(0)"/>
					<format id="colored-debug"  format="%Time %EscM(45)%Level%EscM(49) - %File:%Line - %Msg%n%EscM(0)"/>
					<format id="colored-info"  format="%Time %EscM(46)%Level%EscM(49) - %File:%Line - %Msg%n%EscM(0)"/>
					<format id="colored-warn"  format="%Time %EscM(43)%Level%EscM(49) - %File:%Line - %Msg%n%EscM(0)"/>
					<format id="colored-error"  format="%Time %EscM(41)%Level%EscM(49) - %File:%Line - %Msg%n%EscM(0)"/>
					<format id="colored-critical"  format="%Time %EscM(41)%Level%EscM(49) - %File:%Line - %Msg%n%EscM(0)"/>
				</formats>
			</seelog>`
		logger, _ = log.LoggerFromConfigAsBytes([]byte(testConfig))
	}
	log.ReplaceLogger(logger)

	nodes := protocol.NewNodes()
	l := logic.NewLogic()
	scheduler := logic.NewScheduler()
	//scheduler.CreateExampleFile()
	//return
	webServer := NewWebServer()
	nodeServer := NewNodeServer()
	wsrouter := websocket.NewRouter()
	wsHandler := &WebsocketHandler{}

	err = inject.Populate(config, nodes, l, nodeServer, webServer, scheduler, wsrouter, wsHandler)
	if err != nil {
		panic(err)
	}

	nodeServer.Start() //start the tcp socket server connecting to nodes
	scheduler.Start()  //start the cron scheduler
	wsHandler.Start()  //initialize websocket router
	webServer.Start()  //start the webserver
}
