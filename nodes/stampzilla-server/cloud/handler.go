package cloud

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models/persons"
)

func (c *Connection) MessageFromCloud(msg *models.Message, p *persons.Person) (json.RawMessage, error) {
	switch msg.Type {
	case "forwarded-request":
		fwd, err := models.ParseForwardedRequest(msg.Body)
		if err != nil {
			return nil, err
		}

		logrus.WithFields(logrus.Fields{
			"from":    fwd.RemoteAddr,
			"service": fwd.Service,
		}).Debug("Received forwared request from cloud")

		node := c.store.GetNodeOfType(fwd.Service)
		if node == nil {
			return nil, fmt.Errorf("no such service")
		}

		return c.sender.Request(node.UUID, "cloud-request", fwd, time.Second*10)
	default:
		logrus.WithFields(logrus.Fields{
			"server": c.conn.RemoteAddr().String(),
			"type":   msg.Type,
		}).Warn("TLS: Received unknown message type from cloud")

		return nil, fmt.Errorf("unknown message type")
	}
}
