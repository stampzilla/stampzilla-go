package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	log "github.com/cihub/seelog"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/encoder"
	"github.com/stampzilla/stampzilla-go/protocol"
)

func WebHandlerGetNodes(enc encoder.Encoder) (int, []byte) {
	return 200, encoder.Must(json.Marshal(nodes.All()))
}

func WebHandlerGetNode(params martini.Params) (int, []byte) {
	if n := nodes.Search(params["id"]); n != nil {
		return 200, encoder.Must(json.Marshal(&n))
	}
	return 404, []byte("{}")
}

func WebHandlerCommandToNode(enc encoder.Encoder, params martini.Params, r *http.Request) (int, []byte) {
	log.Info("Sending command to:", params["id"])
	requestJsonPut, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error(err)
		return 500, []byte("Error")
	}
	log.Info("Command:", string(requestJsonPut))

	node := nodes.Search(params["id"])
	if node == nil {
		log.Debug("NODE: ", node)
		return 404, []byte("Node not found")
	}

	node.Conn().Write(requestJsonPut)
	return 200, encoder.Must(enc.Encode(protocol.Command{Cmd: "testresponse"}))
}

//type Command struct {
//Cmd  string
//Args []string
//}

//func PostNodeState(enc encoder.Encoder, params martini.Params, r *http.Request) (int, []byte) {
//// Create a blocking channel
//nodesConnection[params["id"]].wait = make(chan bool)

//soc, ok := nodesConnection[params["id"]]
//if ok {
////c := Command{}

//c := decodeJson(r)
////err := r.DecodeJsonPayload(&c)

//data, err := json.Marshal(&c)

//_, err = soc.conn.Write(data)
//if err != nil {
//log.Error("Failed write: ", err)
//} else {
//log.Info("Success transport command to: ", params["id"])
//}
//} else {
//log.Error("Failed to transport command to: ", params["id"])
//}

//// Wait for answer or timeout..
//select {
//case <-time.After(5 * time.Second):
//case <-nodesConnection[params["id"]].wait:
//}

//n := nodes.GetByName(params["id"])
//if n == nil {
//return 404, []byte("{}")
//}

////w.WriteJson(&n.State)
//return 200, encoder.Must(json.Marshal(&n.State))
//}

//func decodeJson(r *http.Request) interface{} {

//decoder := json.NewDecoder(r.Body)
//var t interface{}
//err := decoder.Decode(&t)
//if err != nil {
//log.Error(err)
//}
//return t
//}
