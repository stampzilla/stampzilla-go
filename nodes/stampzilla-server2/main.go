//go:generate bash -c "go get -u github.com/rakyll/statik && statik -src ./web/dist -f"
package main

import (
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/models"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/servermain"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/store"

	// Statik for the webserver gui
	_ "github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/statik"
)

func main() {

	config := &models.Config{}
	config.MustLoad()

	store := store.New()
	server := servermain.New(config, store)
	server.Init()
	server.Run()
}
