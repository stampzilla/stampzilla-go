package main

import (
    "encoding/json"
    log "github.com/cihub/seelog"
    "github.com/stamp/go-json-rest"
    "time"
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

type Command struct {
    Cmd  string
    Args []string
}

func PostNodeState(w *rest.ResponseWriter, r *rest.Request) {
    // Create a blocking channel
    NodesWait[r.PathParam("id")] = make(chan bool)

    soc, ok := NodesConnection[r.PathParam("id")]
    if ok {
        c := Command{}

        err := r.DecodeJsonPayload(&c)

        data, err := json.Marshal(&c)

        _, err = soc.Write(data)
        if err != nil {
            log.Error("Failed write: ", err)
        } else {
            log.Info("Success transport command to: ", r.PathParam("id"))
        }
    } else {
        log.Error("Failed to transport command to: ", r.PathParam("id"))
    }

    // Wait for answer or timeout..
    select {
    case <-time.After(5 * time.Second):
    case <-NodesWait[r.PathParam("id")]:
    }

    n, ok := Nodes[r.PathParam("id")]
    if !ok {
        rest.NotFound(w, r)
        return
    }

    w.WriteJson(&n.State)
}
