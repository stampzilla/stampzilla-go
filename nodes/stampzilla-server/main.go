//go:generate bash -c "go get -u github.com/rakyll/statik && cd web && npm run build && cd .. && statik -src ./web/dist -f"
package main

import (
	"fmt"

	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/servermain"
	"github.com/stampzilla/stampzilla-go/pkg/node"

	// Statik for the webserver gui
	_ "github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/statik"
)

func main() {

	config := &models.Config{}
	config.MustLoad()

	if config.Version {
		fmt.Printf("Version is: %s\n", node.Version)
		return
	}

	server := servermain.New(config)
	server.Init()
	server.Run()
}
