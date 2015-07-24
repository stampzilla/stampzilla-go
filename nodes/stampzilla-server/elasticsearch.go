package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	log "github.com/cihub/seelog"
	serverprotocol "github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/protocol"
)

type ElasticSearch struct {
	Config *ServerConfig         `inject:""`
	Nodes  *serverprotocol.Nodes `inject:""`

	StateUpdates chan serverprotocol.Node // JSON encoded state updates from the
}

func NewElasticSearch() *ElasticSearch {
	return &ElasticSearch{
		StateUpdates: make(chan serverprotocol.Node, 20),
	}
}

func (self *ElasticSearch) Start() {
	go self.Worker()
}

func (self *ElasticSearch) Worker() {
	for {
		select {
		case update := <-self.StateUpdates:
			self.pushUpdate(update)
		}
	}
}

// METRIC LOGGER INTERFACE
func (self *ElasticSearch) Log(key string, value interface{}) {
}
func (self *ElasticSearch) Commit(node interface{}) {
	if node, ok := node.(serverprotocol.Node); ok {
		self.StateUpdates <- node
	}
}

// END - METRIC LOGGER INTERFACE

func (self *ElasticSearch) pushUpdate(update serverprotocol.Node) {

	type Item struct {
		Name      string      `json:"name"`
		Uuid      string      `json:"uuid"`
		State     interface{} `json:"state"`
		Timestamp string      `json:"@timestamp"`
		Version   string      `json:"@version"`
	}

	item := &Item{
		Name:      update.Name(),
		Uuid:      update.Uuid(),
		State:     update.State(),
		Timestamp: time.Now().UTC().Format("2006-01-02T15:04:05") + ".000Z",
		Version:   "1",
	}

	jsonStr, err := json.Marshal(item)
	if err != nil {
		log.Warn(err)
		return
	}

	url := self.Config.ElasticSearch + "/" + update.Name()
	//log.Debug("URL:>", url)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Warn("Failed to post:", err)
	}
	defer resp.Body.Close()

	//log.Debug("request Body:", string(jsonStr))
	//log.Debug("response Status:", resp.Status)
	//log.Debug("response Headers:", resp.Header)
	ioutil.ReadAll(resp.Body)
	// body, _ :=
	//log.Debug("response Body:", string(body))

}
