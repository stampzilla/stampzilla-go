package cloud

import (
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models/persons"
)

func (c *Connection) MessageFromCloud(msg *models.Message, p *persons.Person) (json.RawMessage, error) {
	switch msg.Type {
	default:
		logrus.WithFields(logrus.Fields{
			"server": c.conn.RemoteAddr().String(),
			"type":   msg.Type,
		}).Warn("TLS: Received unknown message type from cloud")

		return nil, fmt.Errorf("unknown message type")
	}
}
