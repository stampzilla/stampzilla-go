package main

import (
	"flag"

	log "github.com/cihub/seelog"
	fbinject "github.com/facebookgo/inject"
	"github.com/stampzilla/stampzilla-go/stampzilla-server/logic"
	"github.com/stampzilla/stampzilla-go/stampzilla-server/protocol"
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

		logger, _ := log.LoggerFromConfigAsBytes([]byte(testConfig))
		log.ReplaceLogger(logger)
	} else {
		log.ReplaceLogger(logger)
	}

	nodes := protocol.NewNodes()

	//Start websocket
	clients := newClients()

	//Start logic
	logic := logic.NewLogic()
	logic.SetNodes(nodes)
	logic.RestoreRulesFromFile("rules.json")

	// Start Servers and provide dependencies
	webServer := NewWebServer()
	nodeServer := NewNodeServer()
	webHandler := NewWebHandler()
	inject(config, clients, logic, nodes, nodeServer, webHandler, webServer)

	nodeServer.Start()
	webServer.Start()
}

func inject(values ...interface{}) {
	err := fbinject.Populate(values...)
	if err != nil {
		panic(err)
	}
}
