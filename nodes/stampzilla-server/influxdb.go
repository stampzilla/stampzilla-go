package main

import (
	"fmt"
	"time"

	log "github.com/cihub/seelog"
	client "github.com/influxdb/influxdb/client/v2"
	serverprotocol "github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/protocol"
)

type InfluxDb struct {
	Config *ServerConfig         `inject:""`
	Nodes  *serverprotocol.Nodes `inject:""`
	conn   client.Client
}

func NewInfluxDb() *InfluxDb {
	return &InfluxDb{}
}

func (i *InfluxDb) Connect() client.Client {

	conf := client.HTTPConfig{
		Addr:     fmt.Sprintf("http://%s:8086", i.Config.InfluxDbServer),
		Username: i.Config.InfluxDbUser,
		Password: i.Config.InfluxDbPassword,
	}

	con, err := client.NewHTTPClient(conf)
	if err != nil {
		log.Error(err)
	}

	log.Infof("Connected to influxdb: %s", conf.Addr)
	return con
}

func (self *InfluxDb) Start() {
	self.conn = self.Connect()
}

func (self *InfluxDb) Log(key string, value interface{}) {
	log.Trace("Logging: ", key, " = ", value)

	pt, err := client.NewPoint(
		key,
		nil,
		map[string]interface{}{
			"value": value,
		},
		time.Now(),
	)
	if err != nil {
		log.Error(err)
		return
	}

	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  "stampzilla",
		Precision: "s",
	})
	if err != nil {
		log.Error(err)
		return
	}

	bp.AddPoint(pt)
	err = self.conn.Write(bp)
	if err != nil {
		log.Error(err)
	}
}
func (self *InfluxDb) Commit(s interface{}) {
}
