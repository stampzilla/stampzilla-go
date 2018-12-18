package main

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/models"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/servermain"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/store"
)

func setupServer(t *testing.T) (*servermain.Main, func()) {
	config := &models.Config{
		UUID: "123",
		Name: "testserver",
	}
	store := store.New()
	server := servermain.New(config, store)

	prevDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir, err := ioutil.TempDir("", "e2etest")
	if err != nil {
		log.Fatal(err)
	}
	os.Chdir(dir)

	server.Init()
	server.HTTPServer.Init()

	cleanUp := func() {
		os.Chdir(prevDir)
		err := os.RemoveAll(dir) // clean up
		if err != nil {
			t.Fatal(err)
		}
	}
	return server, cleanUp
}
