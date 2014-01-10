package main

import (
    log "github.com/cihub/seelog"
)

func main() {
    // Load logger
    logger, err := log.LoggerFromConfigAsFile("logconfig.xml")
    if err != nil {
        panic(err)
    }
    log.ReplaceLogger(logger)

    log.Info("Starting NET (:8282)")
    netStart()

    log.Info("Starting WEB")
    webStart()

    select {}
}
