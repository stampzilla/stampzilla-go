package main

import (
	"fmt"
	"net/url"
	"time"

	log "github.com/cihub/seelog"
	"github.com/influxdb/influxdb/client"
	serverprotocol "github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/protocol"
)

type InfluxDb struct {
	Config *ServerConfig         `inject:""`
	Nodes  *serverprotocol.Nodes `inject:""`
	conn   *client.Client
}

func NewInfluxDb() *InfluxDb {
	return &InfluxDb{}
}

func (i *InfluxDb) Connect() *client.Client {

	u, err := url.Parse(fmt.Sprintf("http://%s:8086", i.Config.InfluxDbServer))
	if err != nil {
		log.Error(err)
	}

	conf := client.Config{
		URL:      *u,
		Username: i.Config.InfluxDbUser,
		Password: i.Config.InfluxDbPassword,
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
	self.conn = self.Connect()
}

func (self *InfluxDb) Log(key string, value interface{}) {
	var pts = make([]client.Point, 1)
	log.Trace("Logging: ", key, " = ", value)
	pts[0] = client.Point{
		Measurement: key,
		//Tags: map[string]string{
		//"color": "test",
		//"shape": "test",
		//},
		Fields: map[string]interface{}{
			"value": value,
		},
		Time: time.Now(),
		//Precision: "s",
	}

	bps := client.BatchPoints{
		Points:   pts,
		Database: "stampzilla",
	}
	_, err := self.conn.Write(bps)
	if err != nil {
		log.Error(err)
	}
}
func (self *InfluxDb) Commit(s interface{}) {
}
