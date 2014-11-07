package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	log "github.com/cihub/seelog"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/encoder"
	"github.com/stampzilla/stampzilla-go/protocol"
	"github.com/stampzilla/stampzilla-go/stampzilla-server/logic"
	serverprotocol "github.com/stampzilla/stampzilla-go/stampzilla-server/protocol"
)

type WebHandler struct {
	Logic *logic.Logic          `inject:""`
	Nodes *serverprotocol.Nodes `inject:""`
}

func (wh *WebHandler) GetNodes(enc encoder.Encoder) (int, []byte) {
	return 200, encoder.Must(json.Marshal(wh.Nodes.All()))
}

func (wh *WebHandler) GetNode(params martini.Params) (int, []byte) {
	if n := wh.Nodes.Search(params["id"]); n != nil {
		return 200, encoder.Must(json.Marshal(&n))
	}
	return 404, []byte("{}")
}

func (wh *WebHandler) CommandToNodePut(enc encoder.Encoder, params martini.Params, r *http.Request) (int, []byte) {
	log.Info("Sending command to:", params["id"])
	requestJsonPut, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error(err)
		return 500, []byte("Error")
	}
	log.Info("Command:", string(requestJsonPut))

	node := wh.Nodes.Search(params["id"])
	if node == nil {
		log.Debug("NODE: ", node)
		return 404, []byte("Node not found")
	}

	node.Conn().Write(requestJsonPut)
	return 200, encoder.Must(enc.Encode(protocol.Command{Cmd: "testresponse"}))
}

func (wh *WebHandler) CommandToNodeGet(enc encoder.Encoder, params martini.Params, r *http.Request) (int, []byte) {
	node := wh.Nodes.Search(params["id"])
	if node == nil {
		log.Debug("NODE: ", node)
		return 404, []byte("Node not found")
	}

	log.Info("Sending command to:", params["id"])

	//TODO explode _1 on / and generate set attr aswell
	cmd := protocol.Command{Cmd: params["_1"]}

	jsonCmd, err := json.Marshal(cmd)
	if err != nil {
		log.Error(err)
		return 500, []byte("Failed json marshal")
	}
	log.Info("Command:", string(jsonCmd))
	node.Conn().Write(jsonCmd)
	return 200, encoder.Must(enc.Encode(protocol.Command{Cmd: "testresponse"}))
}

func (wh *WebHandler) GetRules() (int, []byte) {
	return 200, encoder.Must(json.Marshal(wh.Logic.Rules()))
}
