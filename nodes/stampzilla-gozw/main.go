package main

import (
	"fmt"

	"github.com/comail/colog"
	"github.com/davecgh/go-spew/spew"
	"github.com/gozwave/gozw/cc"
)

func init() {
	colog.Register()
	colog.ParseFields(true)
}

func main() {
	node := node{}

	appLayer := connectToController()
	node.controller = appLayer

	node.Start()

	node.state.HomeID = fmt.Sprintf("%08X", appLayer.Controller.HomeID)
	node.state.APIVersion = appLayer.Controller.APIVersion
	node.state.Library = appLayer.Controller.APILibraryType
	node.state.Version = fmt.Sprintf("%X", appLayer.Controller.Version)
	node.state.APIType = appLayer.Controller.APIType
	node.state.IsPrimaryController = appLayer.Controller.IsPrimaryController

	for _, v := range appLayer.Nodes() {
		node.state.sync(v.NodeID, *v)
	}

	appLayer.EventBus.Subscribe("node:command", func(nodeID uint8, event cc.Command) {
		fmt.Printf("Node command event n=%d\n", nodeID)
		spew.Dump(event)
	})

	appLayer.EventBus.Subscribe("node:updated", node.state.sync)

	select {}
}
