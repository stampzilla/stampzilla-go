package cloud

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models/devices"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models/persons"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/store"
)

func (c *Connection) SendUpdate(area string, store *store.Store) error {
	c.RLock()
	_, ok := c.subscriptions[area]
	c.RUnlock()

	if !ok {
		return nil
	}

	data, ok := c.store.Get(area)
	if !ok {
		return nil
	}

	message, err := models.NewMessage(area, data)
	if err != nil {
		return err
	}

	_, err = message.WriteToWriter(c.conn)
	if err != nil {
		return err
	}

	return nil
}

func (c *Connection) MessageFromCloud(msg *models.Message, p *persons.Person) (json.RawMessage, error) {
	switch msg.Type {
	case "authorize-request":
		req := &models.AuthorizeRequest{}
		err := json.Unmarshal(msg.Body, req)
		if err != nil {
			return nil, err
		}

		logrus.WithFields(logrus.Fields{
			"username": req.Username,
		}).Debug("Received authorization request from cloud")

		user, err := c.store.ValidateLogin(req.Username, req.Password)
		if user == nil {
			return nil, err
		}

		encoded, err := json.Marshal(user.UUID)
		if err != nil {
			return nil, err
		}

		return encoded, err
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
	case "subscribe":
		subscribeTo := []string{}
		err := json.Unmarshal(msg.Body, &subscribeTo)
		if err != nil {
			return nil, err
		}

		subscriptions := make(map[string]struct{})
		for _, v := range subscribeTo {
			subscriptions[v] = struct{}{}
		}

		c.Lock()
		c.subscriptions = subscriptions
		c.Unlock()

		logrus.Debug("Cloud subscriptions: ", subscribeTo)

		for _, v := range subscribeTo {
			err := c.SendUpdate(v, c.store)
			if err != nil {
				return nil, err
			}
		}

		return nil, nil
	case "state-change":
		devs := devices.NewList()
		err := json.Unmarshal(msg.Body, devs)
		if err != nil {
			return nil, err
		}

		logrus.WithFields(logrus.Fields{
			"from":    msg.FromUUID,
			"devices": devs,
		}).Debug("Received state change request")

		for node, devices := range devs.StateGroupedByNode() {
			logrus.WithFields(logrus.Fields{
				"to": node,
			}).Debug("Send state change request to node")
			c.sender.SendToID(node, "state-change", devices)
		}

		return nil, nil
	default:
		logrus.WithFields(logrus.Fields{
			"server": c.conn.RemoteAddr().String(),
			"type":   msg.Type,
		}).Warn("TLS: Received unknown message type from cloud")

		return nil, fmt.Errorf("unknown message type")
	}
}
