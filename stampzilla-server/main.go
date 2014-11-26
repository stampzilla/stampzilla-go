package main

import (
	"flag"

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

	// Load logger
	logger, err := log.LoggerFromConfigAsFile("logconfig.xml")
	if err != nil {
		testConfig := `
        <seelog type="sync">
            <outputs formatid="main">
                <console/>
            </outputs>
            <formats>
                <format id="main" format="%Ns [%Level] %Msg%n"/>
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
