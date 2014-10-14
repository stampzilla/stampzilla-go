package main

import (
	"errors"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

type Logic struct {
	stateMap map[string]string
	rules    map[string]string
	re       *regexp.Regexp
	sync.RWMutex
}

func NewLogic() *Logic {
	l := &Logic{stateMap: make(map[string]string), rules: make(map[string]string)}
	l.re = regexp.MustCompile("^([^0-9\\s\\[][^\\s\\[]*)?(\\[[0-9]+\\])?$")
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

func (l *Logic) path(state interface{}, jp string, t interface{}) error {
	var v interface{}
	// state should already be unmarshaled here which is done in newclient
	//err := json.Unmarshal(]byte(b), &v)
	//if err != nil {
	//return err
	//}
	if jp == "" {
		return errors.New("invalid path")
	}
	for _, token := range strings.Split(jp, ".") {
		sl := l.re.FindAllStringSubmatch(token, -1)
		if len(sl) == 0 {
			return errors.New("invalid path")
		}
		ss := sl[0]
		if ss[1] != "" {
			v = v.(map[string]interface{})[ss[1]]
		}
		if ss[2] != "" {
			i, err := strconv.Atoi(ss[2][1 : len(ss[2])-1])
			if err != nil {
				return errors.New("invalid path")
			}
			v = v.([]interface{})[i]
		}
	}
	rt := reflect.ValueOf(t).Elem()
	rv := reflect.ValueOf(v)
	rt.Set(rv)
	return nil
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
