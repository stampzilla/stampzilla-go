package main

import (
    //   log "github.com/cihub/seelog"
    "github.com/stamp/go-json-rest"
)

var Nodes map[string]InfoStruct

func GetNodes(w *rest.ResponseWriter, req *rest.Request) {
    w.WriteJson(&Nodes)
}
