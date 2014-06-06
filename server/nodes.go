package main

import (
	"encoding/json"
	log "github.com/cihub/seelog"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/encoder"
	"net/http"
	"time"
)

var Nodes map[string]InfoStruct

func GetNodes(enc encoder.Encoder) (int, []byte) {
	nodes := []InfoStruct{}

	for _, node := range Nodes {
		nodes = append(nodes, node)
	}

	return 200, encoder.Must(json.Marshal(&nodes))
}

func GetNode(enc encoder.Encoder, params martini.Params) (int, []byte) {
	n, ok := Nodes[params["id"]]
	if !ok {
		return 404, []byte("{}")

	}

	return 200, encoder.Must(json.Marshal(&n))
}

type Command struct {
	Cmd  string
	Args []string
}

func PostNodeState(enc encoder.Encoder, params martini.Params, r *http.Request) (int, []byte) {
	// Create a blocking channel
	NodesWait[params["id"]] = make(chan bool)

	soc, ok := NodesConnection[params["id"]]
	if ok {
		//c := Command{}

		c := decodeJson(r)
		//err := r.DecodeJsonPayload(&c)

		data, err := json.Marshal(&c)

		_, err = soc.Write(data)
		if err != nil {
			log.Error("Failed write: ", err)
		} else {
			log.Info("Success transport command to: ", params["id"])
		}
	} else {
		log.Error("Failed to transport command to: ", params["id"])
	}

	// Wait for answer or timeout..
	select {
	case <-time.After(5 * time.Second):
	case <-NodesWait[params["id"]]:
	}

	n, ok := Nodes[params["id"]]
	if !ok {
		return 404, []byte("{}")
	}

	//w.WriteJson(&n.State)
	return 200, encoder.Must(json.Marshal(&n.State))
}

func decodeJson(r *http.Request) interface{} {

	decoder := json.NewDecoder(r.Body)
	var t interface{}
	err := decoder.Decode(&t)
	if err != nil {
		log.Error(err)
	}
	return t
}
