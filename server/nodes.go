package main

import (
    //   log "github.com/cihub/seelog"
    "github.com/stamp/go-json-rest"
)

var Nodes map[string]InfoStruct

func GetNodes(w *rest.ResponseWriter, r *rest.Request) {
    nodes := []InfoStruct{}

    for _, node := range Nodes {
        nodes = append(nodes, node)
    }

    w.WriteJson(&nodes)
}

func GetNode(w *rest.ResponseWriter, r *rest.Request) {
    n, ok := Nodes[r.PathParam("id")]
    if !ok {
        rest.NotFound(w, r)
        return
    }

    w.WriteJson(&n)
}
