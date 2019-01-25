//go:generate bash -c "go get -u github.com/rakyll/statik && statik -src ./web/dist -f"
package main

import (
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/servermain"

	// Statik for the webserver gui
	_ "github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/statik"
)

func main() {

	config := &models.Config{}
	config.MustLoad()

	server := servermain.New(config)
	server.Init()
	server.Run()
}
