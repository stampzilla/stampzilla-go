//go:generate bash -c "go get -u github.com/rakyll/statik && cd web && rm -rf dist && npm run build && cd .. && statik -src ./web/dist -f"
package main

import (
	"fmt"

	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/models"
	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/servermain"

	// Statik for the webserver gui.
	_ "github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/statik"
	"github.com/stampzilla/stampzilla-go/v2/pkg/build"
)

func main() {
	config := &models.Config{}
	config.MustLoad()

	if config.Version {
		fmt.Println(build.String())
		return
	}

	server := servermain.New(config)
	server.Init()
	server.Run()
}
