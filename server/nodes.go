package main

import (
	"encoding/json"
	"net/http"
	"time"

	log "github.com/cihub/seelog"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/encoder"
	"github.com/stampzilla/stampzilla-go/protocol"
)

var Nodes map[string]protocol.Node

func GetNodes(enc encoder.Encoder) (int, []byte) {
	//nodes := []InfoStruct{}

	//for _, node := range Nodes {
	//nodes = append(nodes, node)
	//}

	//return 200, encoder.Must(json.Marshal(&nodes))
	return 200, encoder.Must(json.Marshal(&Nodes))
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
	nodesConnection[params["id"]].wait = make(chan bool)

	soc, ok := nodesConnection[params["id"]]
	if ok {
		//c := Command{}

		c := decodeJson(r)
		//err := r.DecodeJsonPayload(&c)

		data, err := json.Marshal(&c)

		_, err = soc.conn.Write(data)
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
	case <-nodesConnection[params["id"]].wait:
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

func CommandToNode(r *http.Request) {
	//  TODO: implement command here (jonaz) <Fri 03 Oct 2014 05:55:52 PM CEST>
}
