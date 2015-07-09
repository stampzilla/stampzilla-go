package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	log "github.com/cihub/seelog"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/encoder"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/logic"
	serverprotocol "github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/protocol"
	"github.com/stampzilla/stampzilla-go/protocol"
)

type WebHandler struct {
	Logic      *logic.Logic          `inject:""`
	Scheduler  *logic.Scheduler      `inject:""`
	Nodes      *serverprotocol.Nodes `inject:""`
	NodeServer *NodeServer           `inject:""`
}

func (wh *WebHandler) GetNodes(enc encoder.Encoder) (int, []byte) {
	//return 200, encoder.Must(json.Marshal(wh.Nodes.All()))
	jsonRet, err := json.Marshal(wh.Nodes.All())
	if err != nil {
		return 500, []byte(err.Error())
	}
	return 200, jsonRet
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

	node.Write(requestJsonPut)
	return 200, encoder.Must(enc.Encode(protocol.Command{Cmd: "testresponse"}))
}

func (wh *WebHandler) CommandToNodeGet(enc encoder.Encoder, params martini.Params, r *http.Request) (int, []byte) {
	node := wh.Nodes.Search(params["id"])
	if node == nil {
		log.Debug("NODE: ", node)
		return 404, []byte("Node not found")
	}

	log.Info("Sending command to:", params["id"])

	// Split on / to add arguments
	p := strings.Split(params["_1"], "/")
	cmd := protocol.Command{Cmd: p[0], Args: p[1:]}

	jsonCmd, err := json.Marshal(cmd)
	if err != nil {
		log.Error(err)
		return 500, []byte("Failed json marshal")
	}
	log.Info("Command:", string(jsonCmd))
	node.Write(jsonCmd)
	return 200, encoder.Must(enc.Encode(protocol.Command{Cmd: "testresponse"}))
}

func (wh *WebHandler) GetRules() (int, []byte) {
	return 200, encoder.Must(json.Marshal(wh.Logic.Rules()))
}

func (wh *WebHandler) GetScheduleTasks() (int, []byte) {
	return 200, encoder.Must(json.Marshal(wh.Scheduler.Tasks()))
}
func (wh *WebHandler) GetScheduleEntries() (int, []byte) {
	return 200, encoder.Must(json.Marshal(wh.Scheduler.Cron.Entries()))
}

func (wh *WebHandler) GetReload() (int, []byte) {
	wh.Logic.RestoreRulesFromFile("rules.json")
	return 200, encoder.Must(json.Marshal(wh.Logic.Rules()))
}

func (wh *WebHandler) GetServerTrigger(params martini.Params) (int, []byte) {
	wh.NodeServer.Trigger(params["key"], params["value"])
	return 200, encoder.Must(json.Marshal(wh.NodeServer.State.GetState()))
}

func (wh *WebHandler) GetServerSet(params martini.Params) (int, []byte) {
	wh.NodeServer.Set(params["key"], params["value"])
	return 200, encoder.Must(json.Marshal(wh.NodeServer.State.GetState()))
}
