package main

import (
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/models"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/servermain"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/store"
)

func main() {

	config := &models.Config{}
	config.MustLoad()

	store := store.New()
	server := servermain.New(config, store)
	server.Init()
	server.Run()
}
