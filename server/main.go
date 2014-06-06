package main

import (
	"flag"
	log "github.com/cihub/seelog"
)

var netPort string
var webPort string

func main() {

	flag.StringVar(&netPort, "net-port", "8282", "Stampzilla server port")
	flag.StringVar(&webPort, "web-port", "8080", "Webserver port")
	flag.Parse()

	// Load logger
	logger, err := log.LoggerFromConfigAsFile("logconfig.xml")
	if err != nil {
		panic(err)
	}
	log.ReplaceLogger(logger)

	Nodes = make(map[string]InfoStruct)

	log.Info("Starting NET (:" + netPort + ")")
	netStart(netPort)

	log.Info("Starting WEB (:" + webPort + ")")
	webStart(webPort)

}
