package main

import (
	"fmt"
	"net/url"
	"os"
	"time"

	log "github.com/cihub/seelog"
	"github.com/influxdb/influxdb/client"
	serverprotocol "github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/protocol"
)

type InfluxDb struct {
	Config *ServerConfig         `inject:""`
	Nodes  *serverprotocol.Nodes `inject:""`

	StateUpdates chan *serverprotocol.Node // JSON encoded state updates from the
}

func NewInfluxDb() *InfluxDb {
	return &InfluxDb{
		StateUpdates: make(chan *serverprotocol.Node, 20),
	}
}

func (i *InfluxDb) Connect() *client.Client {

	u, err := url.Parse(fmt.Sprintf("http://%s", i.Config.InfluxDbServer))
	if err != nil {
		log.Error(err)
	}

	conf := client.Config{
		URL:      *u,
		Username: os.Getenv("INFLUX_USER"),
		Password: os.Getenv("INFLUX_PWD"),
	}

	con, err := client.NewClient(conf)
	if err != nil {
		log.Error(err)
	}

	dur, ver, err := con.Ping()
	if err != nil {
		log.Error(err)
	}
	log.Infof("Connected to influxdb: %v, %s", dur, ver)

	return con
}

func (self *InfluxDb) Start() {
	conn := self.Connect()
	go self.Worker(conn)
}

func (self *InfluxDb) Worker(conn *client.Client) {
	for {
		select {
		case update := <-self.StateUpdates:
			self.send(conn, update)
		}
	}
}

func (self *InfluxDb) send(conn *client.Client, update *serverprotocol.Node) {
	var pts = make([]client.Point, 1)
	pts[0] = client.Point{
		Name: "shapes",
		Tags: map[string]string{
			"color": "test",
			"shape": "test",
		},
		Fields: map[string]interface{}{
			"value": 10,
		},
		Time:      time.Now(),
		Precision: "s",
	}

	bps := client.BatchPoints{
		Points:   pts,
		Database: "stampzilla",
	}
	conn.Write(bps)

}
