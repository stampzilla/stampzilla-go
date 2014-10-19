package main

import (
	"flag"

	log "github.com/cihub/seelog"
	fbinject "github.com/facebookgo/inject"
	"github.com/stampzilla/stampzilla-go/stampzilla-server/logic"
	"github.com/stampzilla/stampzilla-go/stampzilla-server/protocol"
)

var netPort string
var webPort string
var webRoot string

var nodes *protocol.Nodes
var logicHandler *logic.Logic

func main() {

	flag.StringVar(&netPort, "net-port", "8282", "Stampzilla server port")
	flag.StringVar(&webPort, "web-port", "8080", "Webserver port")
	flag.StringVar(&webRoot, "web-root", "public", "Webserver root")
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

	nodes = protocol.NewNodes()

	//Start websocket
	clients = newClients()

	//Start logic
	logicHandler = logic.NewLogic()
	logicHandler.SetNodes(nodes)
	logicHandler.RestoreRulesFromFile("rules.json")

	// Start NodeServer and provide dependencies
	log.Info("Starting NodeServer (:" + netPort + ")")
	nodeServer := NewNodeServer()
	nodeServer.Port = netPort
	inject(clients, logicHandler, nodes, nodeServer)

	nodeServer.Start()

	log.Info("Starting WEB (:" + webPort + " in " + webRoot + ")")
	webStart(webPort, webRoot)

}

func inject(values ...interface{}) {
	err := fbinject.Populate(values...)
	if err != nil {
		panic(err)
	}
}
