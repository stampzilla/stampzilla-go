package main

import "sync"

type Logic struct {
	stateMap map[string]string
	rules    map[string]string
	sync.RWMutex
}

func NewLogic() *Logic {
	l := &Logic{stateMap: make(map[string]string), rules: make(map[string]string)}
	return l
}

func (l *Logic) ListenForChanges() chan interface{} {
	//TODO maybe this should be a buffered channel so we dont block on send in netStart/newClient
	c := make(chan interface{})
	go l.listen(c)
	return c
}

// listen will run in a own goroutine and listen to incoming state changes and Parse them
func (l *Logic) listen(c chan interface{}) {
	for state := range c {
		l.ParseState(state)
	}
}

func (l *Logic) ParseState(state interface{}) {
	//TODO parse all nodes.State here and generate something like this:
	// OR we dont use stateMap and only use rules Devices[2].On == true and parse it using jsonpath example below.
	// statemap["Devices[1].State"] = "OFF"
	// this might be usefull: http://play.golang.org/p/JQnry4s6KE
	// http://blog.golang.org/json-and-go
}

/*
Example of state:
State: {
	Devices: {
		1: {
			Id: "1",
			Name: "Dev1",
			State: "OFF",
			Type: ""
		},
		2: {
			Id: "2",
			Name: "Dev2",
			State: "ON",
			Type: ""
		}
	}
}
*/
